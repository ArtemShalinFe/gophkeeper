package server

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/ArtemShalinFe/gophkeeper/internal/config"
	"github.com/ArtemShalinFe/gophkeeper/internal/models"
	"github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

type recordDialer struct {
	lis *bufconn.Listener
}

func (d *recordDialer) bufDialer(context.Context, string) (net.Conn, error) {
	return d.lis.Dial()
}

func NewRecordServiceDialer(t *testing.T, us models.UserStorage, rs models.RecordStorage) (*recordDialer, error) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	log := zap.L()
	s, err := InitServer(rs, us, log, config.NewServerCfg())
	if err != nil {
		t.Fatalf("an occured error when initial grpc server, err: %v", err)
	}
	rsrvc := NewRecordsService(log, rs)
	usrvc := NewUsersService(log, us)

	RegisterUsersServer(s.grpcServer, usrvc)
	RegisterRecordsServer(s.grpcServer, rsrvc)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("server exited with error: %v", err)
		}
	}()

	return &recordDialer{
		lis: lis,
	}, nil
}

func contextWithUserID(ctx context.Context, userID string) context.Context {
	headers := map[string]string{
		userIDHeader: userID,
	}

	return metadata.NewOutgoingContext(ctx, metadata.New(headers))
}

func generateMetadata(mic int) []string {
	ss := make([]string, mic)
	for i := 0; i < mic; i++ {
		ss[i] = fmt.Sprintf("%s:%s", uuid.NewString(), uuid.NewString())
	}
	return ss
}

func generateRecord(t *testing.T, dt models.DataType, a models.RecordData) *models.Record {
	t.Helper()

	mic := 10
	mis, err := models.NewMetadataFromStringArray(generateMetadata(mic))
	if err != nil {
		t.Errorf("an error occured while generating metadata, err: %v", err)
	}

	r, err := models.NewRecord(uuid.NewString(), uuid.NewString(), dt,
		time.Now(), time.Now(), a, mis, false, 1)
	if err != nil {
		t.Errorf("an error occured while generating record, err: %v", err)
	}
	return r
}

func generateAuthRecord(t *testing.T) *models.Record {
	return generateRecord(t, models.AuthType, &models.Auth{
		Login:    uuid.NewString(),
		Password: uuid.NewString(),
	})
}

func generateTextRecord(t *testing.T) *models.Record {
	return generateRecord(t, models.TextType, &models.Text{
		Data: uuid.NewString(),
	})
}

func generateBinaryRecord(t *testing.T) *models.Record {
	return generateRecord(t, models.BinaryType, &models.Binary{
		Data: []byte(uuid.NewString()),
	})
}

func generateCardRecord(t *testing.T) *models.Record {
	return generateRecord(t, models.CardType, &models.Card{
		Number: uuid.NewString(),
		Term:   time.Now(),
		Owner:  uuid.NewString(),
	})
}

func generateRecords(t *testing.T, limit int) []*models.Record {
	var records []*models.Record
	for i := 0; i < limit; i++ {
		switch i % 2 {
		case 0:
			records = append(records, generateAuthRecord(t))
		case 1:
			records = append(records, generateTextRecord(t))
		case 3:
			records = append(records, generateBinaryRecord(t))
		default:
			records = append(records, generateCardRecord(t))
		}
		if len(records) >= models.DefaultLimit {
			break
		}
	}
	return records
}

func TestRecordsService_Get(t *testing.T) {
	ctrl := gomock.NewController(t)

	us := NewMockUserStorage(ctrl)
	udto := userDTO()
	u := user(t)
	us.EXPECT().GetUser(gomock.Any(), udto).AnyTimes().Return(u, nil)

	rs := NewMockRecordStorage(ctrl)

	r1 := generateAuthRecord(t)
	r1pb, err := convRecordToProtobuff(r1)
	if err != nil {
		t.Errorf("an error occured while encode auth record to protobuff, err: %v", err)
	}

	r2 := generateTextRecord(t)
	r2pb, err := convRecordToProtobuff(r2)
	if err != nil {
		t.Errorf("an error occured while encode text record to protobuff, err: %v", err)
	}

	r3 := generateBinaryRecord(t)
	r3pb, err := convRecordToProtobuff(r3)
	if err != nil {
		t.Errorf("an error occured while encode binary record to protobuff, err: %v", err)
	}

	r4 := generateCardRecord(t)
	r4pb, err := convRecordToProtobuff(r4)
	if err != nil {
		t.Errorf("an error occured while encode card record to protobuff, err: %v", err)
	}

	rs.EXPECT().Get(gomock.Any(), u.ID, r1.ID).Return(r1, nil)
	rs.EXPECT().Get(gomock.Any(), u.ID, r2.ID).Return(r2, nil)
	rs.EXPECT().Get(gomock.Any(), u.ID, r3.ID).Return(r3, nil)
	rs.EXPECT().Get(gomock.Any(), u.ID, r4.ID).Return(r4, nil)
	rs.EXPECT().Get(gomock.Any(), u.ID, r1.ID).Return(nil, models.ErrRecordNotFound)

	d, err := NewRecordServiceDialer(t, us, rs)
	if err != nil {
		t.Errorf("an occured error when creating a new dialer, err: %v", err)
	}

	tests := []struct {
		name    string
		userid  string
		request *GetRecordRequest
		want    *GetRecordResponse
		wantErr bool
	}{
		{
			name:   "positive case get auth record",
			userid: u.ID,
			request: &GetRecordRequest{
				Id: r1.ID,
			},
			want: &GetRecordResponse{
				Record: r1pb,
			},
			wantErr: false,
		},
		{
			name:   "positive case get text record",
			userid: u.ID,
			request: &GetRecordRequest{
				Id: r2.ID,
			},
			want: &GetRecordResponse{
				Record: r2pb,
			},
			wantErr: false,
		},
		{
			name:   "positive case get binary record",
			userid: u.ID,
			request: &GetRecordRequest{
				Id: r3.ID,
			},
			want: &GetRecordResponse{
				Record: r3pb,
			},
			wantErr: false,
		},
		{
			name:   "positive case get card record",
			userid: u.ID,
			request: &GetRecordRequest{
				Id: r4.ID,
			},
			want: &GetRecordResponse{
				Record: r4pb,
			},
			wantErr: false,
		},
		{
			name:   "negative case not authorized empty string userid",
			userid: " ",
			request: &GetRecordRequest{
				Id: r4.ID,
			},
			want: &GetRecordResponse{
				Record: r4pb,
			},
			wantErr: true,
		},
		{
			name:   "negative case not authorized",
			userid: "",
			request: &GetRecordRequest{
				Id: r4.ID,
			},
			want: &GetRecordResponse{
				Record: r4pb,
			},
			wantErr: true,
		},
		{
			name:   "negative case storage error",
			userid: u.ID,
			request: &GetRecordRequest{
				Id: r1.ID,
			},
			want: &GetRecordResponse{
				Record: r1pb,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rctx := context.Background()

			conn, err := grpc.DialContext(rctx, "bufnet",
				grpc.WithContextDialer(d.bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Errorf("failed to dial bufnet: %v", err)
			}
			defer conn.Close()

			client := NewRecordsClient(conn)

			if tt.userid != "" {
				rctx = contextWithUserID(context.Background(), tt.userid)
			}
			got, err := client.Get(rctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordsService.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.Record, tt.want.Record) {
				t.Errorf("RecordsService.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecordsService_Add(t *testing.T) {
	ctrl := gomock.NewController(t)

	us := NewMockUserStorage(ctrl)
	udto := userDTO()
	u := user(t)
	us.EXPECT().GetUser(gomock.Any(), udto).AnyTimes().Return(u, nil)

	rs := NewMockRecordStorage(ctrl)

	r1 := generateAuthRecord(t)
	r1pb, err := convRecordToProtobuff(r1)
	if err != nil {
		t.Errorf("an error occured while encode auth record to protobuff, err: %v", err)
	}

	r2 := generateTextRecord(t)
	r2pb, err := convRecordToProtobuff(r2)
	if err != nil {
		t.Errorf("an error occured while encode text record to protobuff, err: %v", err)
	}

	r3 := generateBinaryRecord(t)
	r3pb, err := convRecordToProtobuff(r3)
	if err != nil {
		t.Errorf("an error occured while encode binary record to protobuff, err: %v", err)
	}

	r4 := generateCardRecord(t)
	r4pb, err := convRecordToProtobuff(r4)
	if err != nil {
		t.Errorf("an error occured while encode card record to protobuff, err: %v", err)
	}

	rs.EXPECT().Add(gomock.Any(), u.ID, gomock.Any()).Return(r1, nil)
	rs.EXPECT().Add(gomock.Any(), u.ID, gomock.Any()).Return(r2, nil)
	rs.EXPECT().Add(gomock.Any(), u.ID, gomock.Any()).Return(r3, nil)
	rs.EXPECT().Add(gomock.Any(), u.ID, gomock.Any()).Return(r4, nil)
	rs.EXPECT().Add(gomock.Any(), u.ID, gomock.Any()).Return(nil, errSomethingWentWrong)

	d, err := NewRecordServiceDialer(t, us, rs)
	if err != nil {
		t.Errorf("an occured error when creating a new dialer, err: %v", err)
	}

	tests := []struct {
		name    string
		userid  string
		request *AddRecordRequest
		want    *AddRecordResponse
		wantErr bool
	}{
		{
			name:   "positive case add auth record",
			userid: u.ID,
			request: &AddRecordRequest{
				Record: r1pb,
			},
			want: &AddRecordResponse{
				Id: r1.ID,
			},
			wantErr: false,
		},
		{
			name:   "positive case add text record",
			userid: u.ID,
			request: &AddRecordRequest{
				Record: r2pb,
			},
			want: &AddRecordResponse{
				Id: r2.ID,
			},
			wantErr: false,
		},
		{
			name:   "positive case add binary record",
			userid: u.ID,
			request: &AddRecordRequest{
				Record: r3pb,
			},
			want: &AddRecordResponse{
				Id: r3.ID,
			},
			wantErr: false,
		},
		{
			name:   "positive case add card record",
			userid: u.ID,
			request: &AddRecordRequest{
				Record: r4pb,
			},
			want: &AddRecordResponse{
				Id: r4.ID,
			},
			wantErr: false,
		},
		{
			name:   "negative case not authorized empty string userid",
			userid: " ",
			request: &AddRecordRequest{
				Record: r4pb,
			},
			want: &AddRecordResponse{
				Id: r4.ID,
			},
			wantErr: true,
		},
		{
			name:   "negative case not authorized",
			userid: "",
			request: &AddRecordRequest{
				Record: r4pb,
			},
			want: &AddRecordResponse{
				Id: r4.ID,
			},
			wantErr: true,
		},
		{
			name:   "negative case storage error",
			userid: u.ID,
			request: &AddRecordRequest{
				Record: r4pb,
			},
			want: &AddRecordResponse{
				Id: r4.ID,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rctx := context.Background()

			conn, err := grpc.DialContext(rctx, "bufnet",
				grpc.WithContextDialer(d.bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Errorf("failed to dial bufnet: %v", err)
			}
			defer conn.Close()

			client := NewRecordsClient(conn)

			if tt.userid != "" {
				rctx = contextWithUserID(context.Background(), tt.userid)
			}
			got, err := client.Add(rctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordsService.Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.Id, tt.want.Id) {
				t.Errorf("RecordsService.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecordsService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)

	us := NewMockUserStorage(ctrl)
	udto := userDTO()
	u := user(t)
	us.EXPECT().GetUser(gomock.Any(), udto).AnyTimes().Return(u, nil)

	rs := NewMockRecordStorage(ctrl)

	r1 := generateAuthRecord(t)
	r1pb, err := convRecordToProtobuff(r1)
	if err != nil {
		t.Errorf("an error occured while encode auth record to protobuff, err: %v", err)
	}

	r2 := generateTextRecord(t)
	r2pb, err := convRecordToProtobuff(r2)
	if err != nil {
		t.Errorf("an error occured while encode text record to protobuff, err: %v", err)
	}

	r3 := generateBinaryRecord(t)
	r3pb, err := convRecordToProtobuff(r3)
	if err != nil {
		t.Errorf("an error occured while encode binary record to protobuff, err: %v", err)
	}

	r4 := generateCardRecord(t)
	r4pb, err := convRecordToProtobuff(r4)
	if err != nil {
		t.Errorf("an error occured while encode card record to protobuff, err: %v", err)
	}

	rs.EXPECT().Update(gomock.Any(), u.ID, gomock.Any()).Return(r1, nil)
	rs.EXPECT().Update(gomock.Any(), u.ID, gomock.Any()).Return(r2, nil)
	rs.EXPECT().Update(gomock.Any(), u.ID, gomock.Any()).Return(r3, nil)
	rs.EXPECT().Update(gomock.Any(), u.ID, gomock.Any()).Return(r4, nil)
	rs.EXPECT().Update(gomock.Any(), u.ID, gomock.Any()).Return(nil, errSomethingWentWrong)

	d, err := NewRecordServiceDialer(t, us, rs)
	if err != nil {
		t.Errorf("an occured error when creating a new dialer, err: %v", err)
	}

	tests := []struct {
		name    string
		userid  string
		request *UpdateRecordRequest
		want    *UpdateRecordResponse
		wantErr bool
	}{
		{
			name:   "positive case add auth record",
			userid: u.ID,
			request: &UpdateRecordRequest{
				Record: r1pb,
			},
			want: &UpdateRecordResponse{
				Id: r1.ID,
			},
			wantErr: false,
		},
		{
			name:   "positive case add text record",
			userid: u.ID,
			request: &UpdateRecordRequest{
				Record: r2pb,
			},
			want: &UpdateRecordResponse{
				Id: r2.ID,
			},
			wantErr: false,
		},
		{
			name:   "positive case add binary record",
			userid: u.ID,
			request: &UpdateRecordRequest{
				Record: r3pb,
			},
			want: &UpdateRecordResponse{
				Id: r3.ID,
			},
			wantErr: false,
		},
		{
			name:   "positive case add card record",
			userid: u.ID,
			request: &UpdateRecordRequest{
				Record: r4pb,
			},
			want: &UpdateRecordResponse{
				Id: r4.ID,
			},
			wantErr: false,
		},
		{
			name:   "negative case not authorized empty string userid",
			userid: " ",
			request: &UpdateRecordRequest{
				Record: r4pb,
			},
			want: &UpdateRecordResponse{
				Id: r4.ID,
			},
			wantErr: true,
		},
		{
			name:   "negative case not authorized",
			userid: "",
			request: &UpdateRecordRequest{
				Record: r4pb,
			},
			want: &UpdateRecordResponse{
				Id: r4.ID,
			},
			wantErr: true,
		},
		{
			name:   "negative case storage error",
			userid: u.ID,
			request: &UpdateRecordRequest{
				Record: r4pb,
			},
			want: &UpdateRecordResponse{
				Id: r4.ID,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rctx := context.Background()

			conn, err := grpc.DialContext(rctx, "bufnet",
				grpc.WithContextDialer(d.bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Errorf("failed to dial bufnet: %v", err)
			}
			defer conn.Close()

			client := NewRecordsClient(conn)

			if tt.userid != "" {
				rctx = contextWithUserID(context.Background(), tt.userid)
			}
			got, err := client.Update(rctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordsService.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.Id, tt.want.Id) {
				t.Errorf("RecordsService.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecordsService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)

	us := NewMockUserStorage(ctrl)
	udto := userDTO()
	u := user(t)
	us.EXPECT().GetUser(gomock.Any(), udto).AnyTimes().Return(u, nil)

	rs := NewMockRecordStorage(ctrl)

	r1 := generateAuthRecord(t)
	r2 := generateTextRecord(t)
	r3 := generateBinaryRecord(t)
	r4 := generateCardRecord(t)

	rs.EXPECT().Delete(gomock.Any(), u.ID, r1.ID).Return(nil)
	rs.EXPECT().Delete(gomock.Any(), u.ID, r2.ID).Return(nil)
	rs.EXPECT().Delete(gomock.Any(), u.ID, r3.ID).Return(nil)
	rs.EXPECT().Delete(gomock.Any(), u.ID, r4.ID).Return(nil)
	rs.EXPECT().Delete(gomock.Any(), u.ID, r4.ID).Return(errSomethingWentWrong)

	d, err := NewRecordServiceDialer(t, us, rs)
	if err != nil {
		t.Errorf("an occured error when creating a new dialer, err: %v", err)
	}

	tests := []struct {
		name    string
		userid  string
		request *DeleteRecordRequest
		wantErr bool
	}{
		{
			name:   "positive case add auth record",
			userid: u.ID,
			request: &DeleteRecordRequest{
				Id: r1.ID,
			},
			wantErr: false,
		},
		{
			name:   "positive case add text record",
			userid: u.ID,
			request: &DeleteRecordRequest{
				Id: r2.ID,
			},
			wantErr: false,
		},
		{
			name:   "positive case add binary record",
			userid: u.ID,
			request: &DeleteRecordRequest{
				Id: r3.ID,
			},
			wantErr: false,
		},
		{
			name:   "positive case add card record",
			userid: u.ID,
			request: &DeleteRecordRequest{
				Id: r4.ID,
			},
			wantErr: false,
		},
		{
			name:   "negative case not authorized empty string userid",
			userid: " ",
			request: &DeleteRecordRequest{
				Id: r4.ID,
			},
			wantErr: true,
		},
		{
			name:   "negative case not authorized",
			userid: "",
			request: &DeleteRecordRequest{
				Id: r4.ID,
			},
			wantErr: true,
		},
		{
			name:   "negative case storage error",
			userid: u.ID,
			request: &DeleteRecordRequest{
				Id: r4.ID,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rctx := context.Background()

			conn, err := grpc.DialContext(rctx, "bufnet",
				grpc.WithContextDialer(d.bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Errorf("failed to dial bufnet: %v", err)
			}
			defer conn.Close()

			client := NewRecordsClient(conn)

			if tt.userid != "" {
				rctx = contextWithUserID(context.Background(), tt.userid)
			}
			_, err = client.Delete(rctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordsService.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRecordsService_List(t *testing.T) {
	ctrl := gomock.NewController(t)

	us := NewMockUserStorage(ctrl)
	udto := userDTO()
	u := user(t)
	us.EXPECT().GetUser(gomock.Any(), udto).AnyTimes().Return(u, nil)

	rs := NewMockRecordStorage(ctrl)

	d, err := NewRecordServiceDialer(t, us, rs)
	if err != nil {
		t.Errorf("an occured error when creating a new dialer, err: %v", err)
	}
	records := generateRecords(t, models.DefaultLimit)
	rs.EXPECT().List(gomock.Any(), u.ID, 0, models.DefaultLimit).Return(records, nil)

	records2 := generateRecords(t, 7)
	rs.EXPECT().List(gomock.Any(), u.ID, 1, models.DefaultLimit).Return(records2, nil)
	rs.EXPECT().List(gomock.Any(), u.ID, 2, models.DefaultLimit).Return(nil, errSomethingWentWrong)

	tests := []struct {
		name      string
		userid    string
		request   *ListRecordRequest
		wantErr   bool
		wantCount int
	}{
		{
			name:   "positive case list records",
			userid: u.ID,
			request: &ListRecordRequest{
				Offset: 0,
				Limit:  models.DefaultLimit,
			},
			wantErr:   false,
			wantCount: models.DefaultLimit,
		},
		{
			name:   "positive case list records (not full page)",
			userid: u.ID,
			request: &ListRecordRequest{
				Offset: 1,
				Limit:  models.DefaultLimit,
			},
			wantErr:   false,
			wantCount: 7,
		},
		{
			name:   "positive case list records (server return error)",
			userid: u.ID,
			request: &ListRecordRequest{
				Offset: 2,
				Limit:  models.DefaultLimit,
			},
			wantErr:   true,
			wantCount: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rctx := context.Background()

			conn, err := grpc.DialContext(rctx, "bufnet",
				grpc.WithContextDialer(d.bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Errorf("failed to dial bufnet: %v", err)
			}
			defer conn.Close()

			client := NewRecordsClient(conn)

			if tt.userid != "" {
				rctx = contextWithUserID(context.Background(), tt.userid)
			}
			got, err := client.List(rctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordsService.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && (len(got.Records) != tt.wantCount) {
				t.Errorf("RecordsService.List() got = %v, want %v", len(got.Records), tt.wantCount)
				return
			}
		})
	}
}
