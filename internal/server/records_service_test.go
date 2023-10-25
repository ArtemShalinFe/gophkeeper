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

func generateMetainfo(mic int) []string {
	ss := make([]string, mic)
	for i := 0; i < mic; i++ {
		ss[i] = fmt.Sprintf("%s:%s", uuid.NewString(), uuid.NewString())
	}
	return ss
}

func generateRecord(t *testing.T, dt models.DataType, a models.Byter) *models.Record {
	t.Helper()

	mic := 10
	mis, err := models.NewMetainfoFromStringArray(generateMetainfo(mic))
	if err != nil {
		t.Errorf("an error occured while generating metainfo, err: %v", err)
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

func TestRecordsService_Get(t *testing.T) {
	ctrl := gomock.NewController(t)

	us := NewMockUserStorage(ctrl)
	udto := userDTO()
	u := user()
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
			if (got.Error != "") != tt.wantErr {
				t.Errorf("RecordsService.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.Record, tt.want.Record) {
				t.Errorf("RecordsService.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
