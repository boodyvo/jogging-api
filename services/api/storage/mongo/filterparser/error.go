package filterparser

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrIncorrectParenthesisQuery = status.Error(codes.InvalidArgument, "incorrect parenthesis query")
	ErrEmptyQuery                = status.Error(codes.InvalidArgument, "empty query")
	ErrInvalidExpression         = status.Error(codes.InvalidArgument, "invalid expression")
	ErrInvalidValue              = status.Error(codes.InvalidArgument, "invalid value")
	ErrUnknownOperand            = status.Error(codes.InvalidArgument, "unknown operand")
	ErrUnknownTerm               = status.Error(codes.InvalidArgument, "unknown term")
)
