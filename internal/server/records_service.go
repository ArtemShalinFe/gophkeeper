package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"go.uber.org/zap"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ArtemShalinFe/gophkeeper/internal/models"
)

const userIDHeader = "userid"

// RecordsService - Implements GRPC server methods that are responsible for working with user records in storage.
type RecordsService struct {
	UnimplementedRecordsServer
	log           *zap.Logger
	recordStorage models.RecordStorage
}

// NewRecordsService - Object Constructor.
func NewRecordsService(log *zap.Logger, recordStorage models.RecordStorage) *RecordsService {
	return &RecordsService{
		log:           log,
		recordStorage: recordStorage,
	}
}

var errUnauthenticatedTemplate = "an occured error while retrieving user id from context, err: %v"

// GetRecord - used to retrieving record.
func (rs *RecordsService) GetRecord(ctx context.Context, request *GetRecordRequest) (*GetRecordResponse, error) {
	var rr GetRecordResponse

	uid, err := getUserIDFromContext(ctx)
	if err != nil {
		return &rr, status.Errorf(codes.Unauthenticated, fmt.Sprintf(errUnauthenticatedTemplate, err))
	}

	r, err := rs.recordStorage.GetRecord(ctx, uid, request.GetId())
	if err != nil {
		if errors.Is(err, models.ErrRecordNotFound) {
			return &rr, status.Errorf(codes.NotFound, models.ErrRecordNotFound.Error())
		}
		return &rr, status.Errorf(codes.Internal,
			fmt.Sprintf("an occured error while retrieving record from storage, err: %v", err))
	}

	record, err := convRecordToProtobuff(r)
	if err != nil {
		return &rr, status.Errorf(codes.Internal,
			fmt.Sprintf("an occured error while decode record from request, err: %v", err))
	}

	rr.Record = record

	return &rr, nil
}

// AddRecord - add new record to the storage.
func (rs *RecordsService) AddRecord(ctx context.Context, request *AddRecordRequest) (*AddRecordResponse, error) {
	var rr AddRecordResponse

	uid, err := getUserIDFromContext(ctx)
	if err != nil {
		return &rr, status.Errorf(codes.Unauthenticated, fmt.Sprintf(errUnauthenticatedTemplate, err))
	}

	d, err := convDataRecordFromProtobuff(request.Record.Data)
	if err != nil {
		return &rr, status.Errorf(codes.Internal,
			fmt.Sprintf("an error occurred while retrieving a record data from request, err: %v", err))
	}

	rdto, err := models.NewRecordDTO(
		request.Record.GetDescription(),
		convDataTypeFromProtobuff(request.Record.GetType()),
		d,
		convMetadataFromProtobuff(request.Record.Metadata),
	)
	if err != nil {
		return &rr, status.Errorf(codes.Internal,
			fmt.Sprintf("an error occurred while encode record dto from request, err: %v", err))
	}

	if len(rdto.Data) > models.MaxFileSize {
		return &rr, status.Errorf(codes.Internal, models.ErrLargeFile)
	}

	r, err := rs.recordStorage.AddRecord(ctx, uid, rdto)
	if err != nil {
		return &rr, status.Errorf(codes.Internal,
			fmt.Sprintf("an error occurred while add record in storage, err: %v", err))
	}

	rr.Id = r.ID
	return &rr, nil
}

// UpdateRecord - Update record to the storage.
func (rs *RecordsService) UpdateRecord(ctx context.Context,
	request *UpdateRecordRequest) (*UpdateRecordResponse, error) {
	var rr UpdateRecordResponse

	uid, err := getUserIDFromContext(ctx)
	if err != nil {
		return &rr, status.Errorf(codes.Unauthenticated, fmt.Sprintf(errUnauthenticatedTemplate, err))
	}

	r, err := convRecordFromProtobuff(request.Record)
	if err != nil {
		return &rr, status.Errorf(codes.Internal,
			fmt.Sprintf("an error occurred while convert record from protobuff, err: %v", err))
	}

	r.Metadata = convMetadataFromProtobuff(request.Record.Metadata)

	_, err = rs.recordStorage.UpdateRecord(ctx, uid, r)
	if err != nil {
		return &rr, status.Errorf(codes.Internal,
			fmt.Sprintf("an error occurred while update record in storage, err: %v", err))
	}

	rr.Id = r.ID
	return &rr, nil
}

// DeleteRecord - mark records as deleted.
func (rs *RecordsService) DeleteRecord(ctx context.Context,
	request *DeleteRecordRequest) (*DeleteRecordResponse, error) {
	var rr DeleteRecordResponse
	uid, err := getUserIDFromContext(ctx)
	if err != nil {
		return &rr, status.Errorf(codes.Unauthenticated, fmt.Sprintf(errUnauthenticatedTemplate, err))
	}

	recordID := request.GetId()

	if err := rs.recordStorage.DeleteRecord(ctx, uid, recordID); err != nil {
		return &rr, status.Errorf(codes.Internal,
			fmt.Sprintf("an error occurred while add or delete record from storage, err: %v", err))
	}

	return &rr, nil
}

// ListRecords - used to retrieving user records.
func (rs *RecordsService) ListRecords(ctx context.Context, request *ListRecordRequest) (*ListRecordResponse, error) {
	var lr ListRecordResponse

	uid, err := getUserIDFromContext(ctx)
	if err != nil {
		return &lr, status.Errorf(codes.Unauthenticated, fmt.Sprintf(errUnauthenticatedTemplate, err))
	}

	rcs, err := rs.recordStorage.ListRecords(ctx, uid, int(request.Offset), int(request.Limit))
	if err != nil {
		return &lr, status.Errorf(codes.Internal,
			fmt.Sprintf("an occured error while retrieving record list from storage, err: %v", err))
	}

	for _, r := range rcs {
		rc, err := convRecordToProtobuff(r)
		if err != nil {
			return &lr, status.Errorf(codes.Internal,
				fmt.Sprintf("an occured error while decode record list from request, err: %v", err))
		}
		lr.Records = append(lr.Records, rc)
	}

	return &lr, nil
}

func convDataRecordFromProtobuff(rData isRecord_Data) (models.RecordData, error) {
	switch d := rData.(type) {
	case *Record_Auth:
		return &models.Auth{
			Login:    d.Auth.GetLogin(),
			Password: d.Auth.GetPwd(),
		}, nil
	case *Record_Text:
		return &models.Text{
			Data: d.Text.GetData(),
		}, nil
	case *Record_Binary:
		return &models.Binary{
			Data: d.Binary.GetData(),
		}, nil
	case *Record_Card:
		return &models.Card{
			Number: d.Card.GetNumber(),
			Term:   d.Card.GetTerm().AsTime(),
			Owner:  d.Card.GetOwner(),
		}, nil
	default:
		return nil, errors.New("unknow record type")
	}
}

func convDataRecordToProtobuff(r *models.Record) (isRecord_Data, error) {
	switch r.Type {
	case string(models.AuthType):
		auth := &models.Auth{}
		if err := cbor.Unmarshal(r.Data, auth); err != nil {
			return nil, fmt.Errorf("an error occured while encode auth from data, err: %w", err)
		}

		a := &Auth{}
		a.Login = auth.Login
		a.Pwd = auth.Password

		return &Record_Auth{Auth: a}, nil
	case string(models.TextType):
		text := &models.Text{}
		if err := cbor.Unmarshal(r.Data, text); err != nil {
			return nil, fmt.Errorf("an error occured while encode text from data, err: %w", err)
		}

		t := &Text{}
		t.Data = text.Data

		return &Record_Text{Text: t}, nil
	case string(models.BinaryType):
		bin := &models.Binary{}
		if err := cbor.Unmarshal(r.Data, bin); err != nil {
			return nil, fmt.Errorf("an error occured while encode binary from data, err: %w", err)
		}

		b := &Binary{}
		b.Data = bin.Data

		return &Record_Binary{Binary: b}, nil
	case string(models.CardType):
		card := &models.Card{}
		if err := cbor.Unmarshal(r.Data, card); err != nil {
			return nil, fmt.Errorf("an error occured while encode card from data, err: %w", err)
		}

		c := &Card{}
		c.Number = card.Number
		c.Term = timestamppb.New(card.Term)
		c.Owner = card.Owner

		return &Record_Card{Card: c}, nil
	default:
		return nil, errors.New("unknow protobuff type")
	}
}

func convRecordFromProtobuff(r *Record) (*models.Record, error) {
	d, err := convDataRecordFromProtobuff(r.Data)
	if err != nil {
		return nil, fmt.Errorf("an error occured while decode record from protobuff, err: %w", err)
	}

	b, err := d.BinData()
	if err != nil {
		return nil, fmt.Errorf("an error occured while encode data record to bytes, err: %w", err)
	}

	return &models.Record{
		ID:          r.GetId(),
		Owner:       r.GetOwner(),
		Description: r.GetDescription(),
		Type:        r.GetType().String(),
		Data:        b,
		Hashsum:     r.GetHashsum(),
		Metadata:    convMetadataFromProtobuff(r.GetMetadata()),
		Version:     r.Version,
	}, nil
}

func convDataTypeFromProtobuff(dt DataType) models.DataType {
	switch dt {
	case *DataType_AUTH.Enum():
		return models.AuthType
	case *DataType_TEXT.Enum():
		return models.TextType
	case *DataType_BINARY.Enum():
		return models.BinaryType
	case *DataType_CARD.Enum():
		return models.CardType
	default:
		return "unknow type"
	}
}

func convDataTypeToProtobuff(t string) DataType {
	switch t {
	case string(models.AuthType):
		return DataType_AUTH
	case string(models.TextType):
		return DataType_TEXT
	case string(models.BinaryType):
		return DataType_BINARY
	case string(models.CardType):
		return DataType_CARD
	default:
		return DataType_UNKNOWN
	}
}

func convRecordToProtobuff(r *models.Record) (*Record, error) {
	data, err := convDataRecordToProtobuff(r)
	if err != nil {
		return nil, fmt.Errorf("an error occured while encode record from protobuff, err: %w", err)
	}

	return &Record{
		Id:          r.ID,
		Owner:       r.Owner,
		Description: r.Description,
		Type:        convDataTypeToProtobuff(r.Type),
		Created:     timestamppb.New(r.Created),
		Modified:    timestamppb.New(r.Modified),
		Data:        data,
		Hashsum:     r.Hashsum,
		Metadata:    convMetadataToProtobuff(r.Metadata),
		Version:     r.Version,
	}, nil
}

func convMetadataFromProtobuff(m []*Metadata) []*models.Metadata {
	mi := make([]*models.Metadata, len(m))
	for i := 0; i < len(m); i++ {
		mi[i] = &models.Metadata{
			Key:   m[i].Key,
			Value: m[i].Value,
		}
	}
	return mi
}

func convMetadataToProtobuff(m []*models.Metadata) []*Metadata {
	mi := make([]*Metadata, len(m))
	for i := 0; i < len(m); i++ {
		mi[i] = &Metadata{
			Key:   m[i].Key,
			Value: m[i].Value,
		}
	}
	return mi
}

func getUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("headers for request is required")
	}

	ids := md.Get(userIDHeader)
	if len(ids) == 0 {
		return "", fmt.Errorf("header %s is required", userIDHeader)
	}

	id := ids[0]

	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("header %s is empty", userIDHeader)
	}

	return id, nil
}
