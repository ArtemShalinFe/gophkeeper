package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	var rr GetRecordResponse

	uid, err := getUserIDFromContext(ctx)
	if err != nil {
		er := fmt.Sprintf(errUnauthenticatedTemplate, err)
		rr.Error = er
		return &rr, status.Errorf(codes.Unauthenticated, er)
	}

	r, err := rs.recordStorage.Get(ctx, uid, request.GetId())
	if err != nil {
		er := fmt.Sprintf("an occured error while retrieving record from storage, err: %v", err)
		rr.Error = er
		return &rr, status.Errorf(codes.NotFound, er)
	}

	record, err := convRecordToProtobuff(r)
	if err != nil {
		er := fmt.Sprintf("an occured error while decode record from request, err: %v", err)
		rr.Error = er
		return &rr, status.Errorf(codes.Internal, er)
	}

	// TODO add decrypt d

	rr.Record = record

	return &rr, nil
}

func (rs *RecordsService) Add(ctx context.Context, request *AddRecordRequest) (*AddRecordResponse, error) {
	var rr AddRecordResponse

	uid, err := getUserIDFromContext(ctx)
	if err != nil {
		er := fmt.Sprintf(errUnauthenticatedTemplate, err)
		rr.Error = er
		return &rr, status.Errorf(codes.Unauthenticated, er)
	}

	d, err := convDataRecordFromProtobuff(request.Record.Data)
	if err != nil {
		er := fmt.Sprintf("an error occurred while retrieving a record data from request, err: %v", err)
		rr.Error = er
		return &rr, status.Errorf(codes.Internal, er)
	}

	// TODO add encrypt b

	rdto, err := models.NewRecordDTO(
		request.Record.GetDescription(),
		convDataTypeFromProtobuff(request.Record.GetType()),
		d,
		convFromPBMetainfo(request.Record.Metainfo),
	)
	if err != nil {
		er := fmt.Sprintf("an error occurred while encode record dto from request, err: %v", err)
		rr.Error = er
		return &rr, status.Errorf(codes.Internal, er)
	}

	r, err := rs.recordStorage.Add(ctx, uid, rdto)
	if err != nil {
		er := fmt.Sprintf("an error occurred while add record in storage, err: %v", err)
		rr.Error = er
		return &rr, status.Errorf(codes.Internal, er)
	}

	rr.Id = r.ID
	return &rr, nil
}

func (rs *RecordsService) Update(ctx context.Context, request *UpdateRecordRequest) (*UpdateRecordResponse, error) {
	var rr UpdateRecordResponse

	uid, err := getUserIDFromContext(ctx)
	if err != nil {
		er := fmt.Sprintf(errUnauthenticatedTemplate, err)
		rr.Error = er
		return &rr, status.Errorf(codes.Unauthenticated, er)
	}

	r, err := convFromProtobuffToRecord(request.Record)
	if err != nil {
		er := fmt.Sprintf("an error occurred while convert record from protobuff, err: %v", err)
		rr.Error = er
		return &rr, status.Errorf(codes.Internal, er)
	}

	r.Metainfo = convFromPBMetainfo(request.Record.Metainfo)

	// TODO add encrypt r.Data bytes

	_, err = rs.recordStorage.Update(ctx, uid, r)
	if err != nil {
		er := fmt.Sprintf("an error occurred while update record in storage, err: %v", err)
		rr.Error = er
		return &rr, status.Errorf(codes.Internal, er)
	}

	rr.Id = r.ID
	return &rr, nil
}

func (rs *RecordsService) Delete(ctx context.Context, request *DeleteRecordRequest) (*DeleteRecordResponse, error) {
	var rr DeleteRecordResponse
	uid, err := getUserIDFromContext(ctx)
	if err != nil {
		er := fmt.Sprintf(errUnauthenticatedTemplate, err)
		rr.Error = er
		return &rr, status.Errorf(codes.Unauthenticated, er)
	}

	recordID := request.GetId()

	if err := rs.recordStorage.Delete(ctx, uid, recordID); err != nil {
		er := fmt.Sprintf("an error occurred while add or delete record from storage, err: %v", err)
		rr.Error = er
		return &rr, status.Errorf(codes.Internal, er)
	}

	return &rr, nil
}

func convDataRecordFromProtobuff(rData isRecord_Data) (models.Byter, error) {
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
		Metainfo:    convMetainfoToProtobuff(r.Metainfo),
	}, nil
}

func convFromPBMetainfo(m []*Metainfo) []*models.Metainfo {
	mi := make([]*models.Metainfo, len(m))
	for i := 0; i < len(m); i++ {
		mi[i] = &models.Metainfo{
			Key:   m[i].Key,
			Value: m[i].Value,
		}
	}
	return mi
}

func convMetainfoToProtobuff(m []*models.Metainfo) []*Metainfo {
	mi := make([]*Metainfo, len(m))
	for i := 0; i < len(m); i++ {
		mi[i] = &Metainfo{
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
