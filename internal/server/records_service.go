package server

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ArtemShalinFe/gophkeeper/internal/models"
)

const userIDHeader = "userid"

type RecordsService struct {
	UnimplementedRecordsServer
	log           *zap.Logger
	recordStorage models.RecordStorage
}

func NewRecordsService(log *zap.Logger, recordStorage models.RecordStorage) *RecordsService {
	return &RecordsService{
		log:           log,
		recordStorage: recordStorage,
	}
}

var errUnauthenticatedTemplate = "an occured error while retrieving user id from context, err: %v"

func (rs *RecordsService) Get(ctx context.Context, request *GetRecordRequest) (*GetRecordResponse, error) {
	uid, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, errUnauthenticatedTemplate, err)
	}

	r, err := rs.recordStorage.Get(ctx, uid, request.GetId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound,
			"an occured error while retrieving record from storage, err: %v", err)
	}

	d, err := convDataRecordToProtobuff(r)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"an occured error while decode data record from request, err: %v", err)
	}

	// TODO add decrypt d

	var rr GetRecordResponse
	rr.Record = &Record{
		Id:       r.ID,
		Data:     d,
		Metainfo: convMetainfoToProtobuff(r.Metainfo),
	}

	return &rr, nil
}

func (rs *RecordsService) Add(ctx context.Context, request *AddRecordRequest) (*AddRecordResponse, error) {
	uid, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, errUnauthenticatedTemplate, err)
	}

	d, err := convDataRecordFromProtobuff(request.Record.Data)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"an error occurred while retrieving a record data from request, err: %v", err)
	}

	b, err := d.ToByte()
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"an error occurred while encode to bytes, err: %v", err)
	}

	// TODO add encrypt b

	var rdto models.RecordDTO
	rdto.Description = request.Record.GetDescription()
	rdto.Hashsum = request.Record.GetHashsum()
	rdto.Data = b
	rdto.Metainfo = convFromPBMetainfo(request.Record.Metainfo)

	r, err := rs.recordStorage.Add(ctx, uid, &rdto)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"an error occurred while add record in storage, err: %v", err)
	}

	var rr AddRecordResponse
	rr.Id = r.ID
	return &rr, nil
}

func (rs *RecordsService) Update(ctx context.Context, request *UpdateRecordRequest) (*UpdateRecordResponse, error) {
	uid, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, errUnauthenticatedTemplate, err)
	}

	r, err := convFromProtobuffToRecord(request.Record)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"an error occurred while convert record from protobuff, err: %v", err)
	}

	r.Metainfo = convFromPBMetainfo(request.Record.Metainfo)

	// TODO add encrypt r.Data bytes

	_, err = rs.recordStorage.Update(ctx, uid, r)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"an error occurred while update record in storage, err: %v", err)
	}

	var rr UpdateRecordResponse
	rr.Id = r.ID
	return &rr, nil
}

func (rs *RecordsService) Delete(ctx context.Context, request *DeleteRecordRequest) (*DeleteRecordResponse, error) {
	uid, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, errUnauthenticatedTemplate, err)
	}

	recordID := request.GetId()

	if err := rs.recordStorage.Delete(ctx, uid, recordID); err != nil {
		return nil, status.Errorf(codes.Internal,
			"an error occurred while add or delete record from storage, err: %v", err)
	}

	return &DeleteRecordResponse{}, nil
}

func convDataRecordFromProtobuff(rData isRecord_Data) (models.Byter, error) {
	switch d := rData.(type) {
	case *Record_Auth:
		return &models.Auth{
			Login:    d.Auth.GetLogin(),
			Password: d.Auth.GetLogin(),
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
		if err := auth.FromByte(r.Data); err != nil {
			return nil, fmt.Errorf("an error occured while encode auth from data, err: %w", err)
		}

		a := &Auth{}
		a.Login = auth.Login
		a.Pwd = auth.Password

		return &Record_Auth{Auth: a}, nil
	case string(models.TextType):
		text := &models.Text{}
		if err := text.FromByte(r.Data); err != nil {
			return nil, fmt.Errorf("an error occured while encode text from data, err: %w", err)
		}

		t := &Text{}
		t.Data = text.Data

		return &Record_Text{Text: t}, nil
	case string(models.BinaryType):
		bin := &models.Binary{}
		if err := bin.FromByte(r.Data); err != nil {
			return nil, fmt.Errorf("an error occured while encode binary from data, err: %w", err)
		}

		b := &Binary{}
		b.Data = bin.Data

		return &Record_Binary{Binary: b}, nil
	case string(models.CardType):
		card := &models.Card{}
		if err := card.FromByte(r.Data); err != nil {
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

func convFromProtobuffToRecord(r *Record) (*models.Record, error) {
	d, err := convDataRecordFromProtobuff(r.Data)
	if err != nil {
		return nil, fmt.Errorf("an error occured while decode record from protobuff, err: %w", err)
	}

	b, err := d.ToByte()
	if err != nil {
		return nil, fmt.Errorf("an error occured while encode record to bytes, err: %w", err)
	}

	return &models.Record{
		ID:          r.GetId(),
		Owner:       r.GetOwner(),
		Description: r.GetDescription(),
		Type:        r.GetType().String(),
		Data:        b,
		Hashsum:     r.GetHashsum(),
		Metainfo:    convFromPBMetainfo(r.GetMetainfo()),
	}, nil
}

func convDataTypeToProtobuff(t string) DataType {
	switch t {
	case DataType_AUTH.String():
		return DataType_AUTH
	case DataType_TEXT.String():
		return DataType_TEXT
	case DataType_BINARY.String():
		return DataType_BINARY
	case DataType_CARD.String():
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
		Metainfo:    convMetainfoToProtobuff(r.Metainfo),
	}, nil
}

func convFromPBMetainfo(m []*Metainfo) []*models.Metainfo {
	mi := make([]*models.Metainfo, len(m))
	for i := 0; i < len(m); i++ {
		mi[i].Key = m[i].Key
		mi[i].Value = m[i].Value
	}
	return mi
}

func convMetainfoToProtobuff(m []*models.Metainfo) []*Metainfo {
	mi := make([]*Metainfo, len(m))
	for i := 0; i < len(m); i++ {
		mi[i].Key = m[i].Key
		mi[i].Value = m[i].Value
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

	return id, nil
}
