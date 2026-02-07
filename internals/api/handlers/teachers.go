package handlers

import (
	"context"

	"school_project_grpc/internals/models"
	"school_project_grpc/internals/repositories"
	"school_project_grpc/pkg/utils"
	pb "school_project_grpc/proto/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Add teachers
func (s *Server) AddTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {

	// Validate: ID must be empty on create
	for _, teacher := range req.Teachers {
		if teacher.Id != "" {
			return nil, status.Error(codes.InvalidArgument,
				"request is incorrect format: non-empty ID fields are not allowed.")
		}
	}

	addedTeacher, err := repositories.AddTeachersDBHandler(ctx, req.GetTeachers())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: addedTeacher}, nil
}

// Get teachers with filter + sort
func (s *Server) GetTeachers(ctx context.Context, req *pb.GetTeacherRequset) (*pb.Teachers, error) {

	// Build Mongo filter from request
	filter, err := buildfilter(req.Teacher, &models.Teacher{})
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal err")
	}

	// Build sort options from request
	sortOption := buildSortOptions(req.GetSortBy())

	pageNumber := req.GetPageNum()
	pageSize := req.GetPageSize()

	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Fetch from database
	teachers, err := repositories.GetTeachersDBhandler(ctx, sortOption, filter, pageSize, pageNumber)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.Teachers{Teachers: teachers}, nil
}

// Update teachers
func (s *Server) UpdateTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {
	updatedTeachers, err := repositories.UpdateTeachersDBHandler(ctx, req.Teachers)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: updatedTeachers}, nil
}

// Delete teachers by IDs
func (s *Server) DeleteTeachers(ctx context.Context, req *pb.TeacherIds) (*pb.DeleteTeacherConfirm, error) {

	ids := req.TeacherIds
	var teacherIDsTODelete []string

	// Collect string IDs
	for _, v := range ids {
		teacherIDsTODelete = append(teacherIDsTODelete, v.Id)
	}

	deletedIds, err := repositories.DeleteTeachersDBHandler(ctx, teacherIDsTODelete)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.DeleteTeacherConfirm{
		Status:     "Teacher successfully deleted",
		DeletedIds: deletedIds,
	}, nil
}

// get students that are asigned to a spacific id
func (s *Server) GetStudentsByClassTeacher(ctx context.Context, req *pb.TeacherId) (*pb.Students, error) {

	// getting the id into variable
	id := req.GetId()

	students, err := repositories.GetStudentByTeacherIDDBhandler(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Students{Students: students}, nil
}

func (s *Server) GetStudentCountByClassTeacher(ctx context.Context, req *pb.TeacherId) (*pb.StudentCount, error) {
	id := req.GetId()

	count, err := repositories.GetStudentCountByTeacherDBHandler(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.StudentCount{
		Status:       true,
		StudentCount: int32(count),
	}, nil
}
