package transport

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"desa-agent/internal/models"
	"desa-agent/internal/usecase"
	pb "desa-agent/pkg/users"
)

type UsersServiceServer struct {
	pb.UnimplementedUsersServiceServer
	uc *usecase.UsersUseCase
}

func NewUsersServiceServer(uc *usecase.UsersUseCase) *UsersServiceServer {
	return &UsersServiceServer{uc: uc}
}

func (s *UsersServiceServer) Register(grpcServer *grpc.Server) {
	pb.RegisterUsersServiceServer(grpcServer, s)
}

func (s *UsersServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	if req.UserHash == "" {
		return nil, status.Error(codes.InvalidArgument, "user_hash is required")
	}

	user, err := s.uc.GetUser(ctx, req.UserHash, req.IncludePii)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return toProtoUser(user), nil
}

func (s *UsersServiceServer) ListUsers(req *pb.ListUsersRequest, stream grpc.ServerStreamingServer[pb.User]) error {
	ctx := stream.Context()

	usersCh, errCh := s.uc.ListUsers(ctx, req.IncludePii)

	for {
		select {
		case <-ctx.Done():
			return status.Error(codes.Canceled, "request canceled")

		case err := <-errCh:
			if err != nil {
				return status.Errorf(codes.Internal, "failed to list users: %v", err)
			}

		case user, ok := <-usersCh:
			if !ok {
				return nil
			}

			if err := stream.Send(toProtoUser(&user)); err != nil {
				return status.Errorf(codes.Internal, "failed to send user: %v", err)
			}
		}
	}
}

func toProtoUser(u *models.User) *pb.User {
	if u == nil {
		return nil
	}

	protoUser := &pb.User{
		UserHash: u.UserHash,
		Status:   toProtoUserStatus(u.Status),
		IdpType:  toProtoIdpType(u.IdpType),
	}

	if u.PII != nil {
		protoUser.UserPii = toProtoUserPII(u.PII)
	}

	return protoUser
}

func toProtoUserPII(pii *models.UserPII) *pb.UserPII {
	if pii == nil {
		return nil
	}

	protoPII := &pb.UserPII{}

	if pii.Username != "" {
		protoPII.Username = &pii.Username
	}
	if pii.Email != "" {
		protoPII.Email = &pii.Email
	}
	if pii.DisplayName != "" {
		protoPII.DisplayName = &pii.DisplayName
	}
	if pii.FirstName != "" {
		protoPII.FirstName = &pii.FirstName
	}
	if pii.LastName != "" {
		protoPII.LastName = &pii.LastName
	}
	if pii.Phone != "" {
		protoPII.Phone = &pii.Phone
	}
	if pii.Department != "" {
		protoPII.Department = &pii.Department
	}
	if pii.Title != "" {
		protoPII.Title = &pii.Title
	}
	if pii.ManagerID != "" {
		protoPII.ManagerId = &pii.ManagerID
	}
	if pii.EmployeeID != "" {
		protoPII.EmployeeId = &pii.EmployeeID
	}
	if pii.Location != "" {
		protoPII.Location = &pii.Location
	}

	for _, attr := range pii.Attributes {
		protoPII.Attributes = append(protoPII.Attributes, &pb.Attribute{
			Key:   pb.AttributeKey(attr.Key),
			Value: attr.Value,
		})
	}

	return protoPII
}

func toProtoUserStatus(status models.UserStatus) pb.UserStatus {
	switch status {
	case models.UserStatusActive:
		return pb.UserStatus_USER_STATUS_ACTIVE
	case models.UserStatusDisabled:
		return pb.UserStatus_USER_STATUS_DISABLED
	default:
		return pb.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

func toProtoIdpType(idpType models.IdentityProviderType) pb.IdentityProviderType {
	switch idpType {
	case models.IdentityProviderTypeActiveDirectory:
		return pb.IdentityProviderType_IDENTITY_PROVIDER_TYPE_ACTIVE_DIRECTORY
	case models.IdentityProviderTypeLDAP:
		return pb.IdentityProviderType_IDENTITY_PROVIDER_TYPE_LDAP
	default:
		return pb.IdentityProviderType_IDENTITY_PROVIDER_TYPE_UNSPECIFIED
	}
}
