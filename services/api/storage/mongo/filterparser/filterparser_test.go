package filterparser

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/boodyvo/jogging-api/lib"
	"gopkg.in/mgo.v2/bson"

	"github.com/stretchr/testify/assert"
)

func TestParseParenthesis(t *testing.T) {
	type TestCase struct {
		Name   string
		Query  string
		Result map[int]int
		Err    error
	}

	tests := []TestCase{
		{
			Name:   "empty",
			Query:  "",
			Result: map[int]int{},
		},
		{
			Name:  "correct",
			Query: "((ex1))(ex2)(ex3)(ex4)and((ex5)(ex6))or((ex7)((ex8)(ex9)))",
			Result: map[int]int{
				0:  6,
				1:  5,
				7:  11,
				12: 16,
				17: 21,
				25: 36,
				26: 30,
				31: 35,
				39: 57,
				40: 44,
				45: 56,
				46: 50,
				51: 55,
			},
		},
		{
			Name:  "incorrect parenthesis order",
			Query: ")(ex1)(",
			Err:   ErrIncorrectParenthesisQuery,
		},
		{
			Name:  "incorrect number of opened parenthesis",
			Query: "(ex1)(",
			Err:   ErrIncorrectParenthesisQuery,
		},
		{
			Name:  "incorrect number of closed parenthesis",
			Query: "(ex1)(ex2))",
			Err:   ErrIncorrectParenthesisQuery,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(tt *testing.T) {
			res, err := parseParenthesis(tc.Query)
			if tc.Err != nil {
				assert.Equal(tt, tc.Err, err, "error is incorrect")

				return
			}

			assert.Equal(tt, tc.Result, res, "result is incorrect")
		})
	}
}

func TestCalcExpressions(t *testing.T) {
	type TestCase struct {
		Name   string
		Exp1   interface{}
		O      string
		Exp2   interface{}
		Result bson.D
		Err    error
	}

	tm, _ := time.Parse(lib.DateFormat, "2020-03-20")
	tests := []TestCase{
		{
			Name: "and",
			Exp1: bson.D{{"$and", []bson.D{
				{{"distance", bson.D{{"$gt", 100}}}},
				{{"date", bson.D{{"$eq", "2020-03-20"}}}},
			}}},
			O: "and",
			Exp2: bson.D{{"$and", []bson.D{
				{{"distance", bson.D{{"$lt", 200}}}},
				{{"date", bson.D{{"$eq", "2020-03-20"}}}},
			}}},
			Result: bson.D{{"$and", []bson.D{
				{{"$and", []bson.D{
					{{"distance", bson.D{{"$gt", 100}}}},
					{{"date", bson.D{{"$eq", "2020-03-20"}}}},
				}}},
				{{"$and", []bson.D{
					{{"distance", bson.D{{"$lt", 200}}}},
					{{"date", bson.D{{"$eq", "2020-03-20"}}}},
				}}},
			}}},
		},
		{
			Name: "or",
			Exp1: bson.D{{"$and", []bson.D{
				{{"distance", bson.D{{"$gt", 100}}}},
				{{"date", bson.D{{"$eq", tm}}}},
			}}},
			O: "or",
			Exp2: bson.D{{"$and", []bson.D{
				{{"distance", bson.D{{"$lt", 200}}}},
				{{"date", bson.D{{"$eq", tm}}}},
			}}},
			Result: bson.D{{"$or", []bson.D{
				{{"$and", []bson.D{
					{{"distance", bson.D{{"$gt", 100}}}},
					{{"date", bson.D{{"$eq", tm}}}},
				}}},
				{{"$and", []bson.D{
					{{"distance", bson.D{{"$lt", 200}}}},
					{{"date", bson.D{{"$eq", tm}}}},
				}}},
			}}},
		},
		{
			Name:   "$gt",
			Exp1:   "distance",
			O:      "gt",
			Exp2:   "100",
			Result: bson.D{{"distance", bson.D{{"$gt", float32(100)}}}},
		},
		{
			Name:   "$lt",
			Exp1:   "distance",
			O:      "lt",
			Exp2:   "100",
			Result: bson.D{{"distance", bson.D{{"$lt", float32(100)}}}},
		},
		{
			Name:   "$eq",
			Exp1:   "distance",
			O:      "eq",
			Exp2:   "100",
			Result: bson.D{{"distance", bson.D{{"$eq", float32(100)}}}},
		},
		{
			Name:   "$ne",
			Exp1:   "distance",
			O:      "ne",
			Exp2:   "100",
			Result: bson.D{{"distance", bson.D{{"$ne", float32(100)}}}},
		},
		{
			Name: "invalid value",
			Exp1: "distance",
			O:    "ne",
			Exp2: "12enqwu9q",
			Err:  ErrInvalidValue,
		},
		{
			Name: "unknown operand",
			Exp1: "distance",
			O:    "nasdae",
			Exp2: "12enqwu9q",
			Err:  ErrUnknownOperand,
		},
		{
			Name: "invalid expression",
			Exp1: []string{"asdasdad"},
			O:    "and",
			Exp2: "12enqwu9q",
			Err:  ErrInvalidExpression,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(tt *testing.T) {
			res, err := calcExpressions(tc.Exp1, tc.O, tc.Exp2, termsTracking)
			if tc.Err != nil {
				assert.Equal(tt, tc.Err, err, "error is incorrect")

				return
			}

			assert.Equal(
				tt,
				true,
				reflect.DeepEqual(tc.Result, res),
				fmt.Sprintf("result is incorrect:\nexpected: %v\nactual: %v\n", tc.Result, res),
			)
		})
	}
}

func TestParseTracking(t *testing.T) {
	type TestCase struct {
		Name   string
		Query  string
		Result bson.D
		Err    error
	}

	tm, _ := time.Parse(lib.DateFormat, "2020-03-22")
	tests := []TestCase{
		{
			Name:  "two parenthesis and expression",
			Query: "(distance gt 100) and (date eq 2020-03-22)",
			Result: bson.D{{"$and", []bson.D{
				{{"distance", bson.D{{"$gt", float32(100)}}}},
				{{"date", bson.D{{"$eq", tm}}}},
			}}},
		},
		{
			Name:  "nested expression",
			Query: "((distance gt 100) or (time lt 10000s)) and (date eq 2020-03-22)",
			Result: bson.D{{"$and", []bson.D{
				{{"$or", []bson.D{
					{{"distance", bson.D{{"$gt", float32(100)}}}},
					{{"time", bson.D{{"$lt", time.Duration(10000 * time.Second)}}}},
				}}},
				{{"date", bson.D{{"$eq", tm}}}},
			}}},
		},
		{
			Name:  "incorrect parenthesis",
			Query: "((distance gt 100 or (time lt 10000s)) and (date eq 2020-03-22)",
			Err:   ErrIncorrectParenthesisQuery,
		},
		{
			Name:  "incorrect expression",
			Query: "(distance gt 100 or time)",
			Err:   ErrInvalidExpression,
		},
		{
			Name:  "unknown term",
			Query: "(100 or time)",
			Err:   ErrUnknownTerm,
		},
		{
			Name:  "correct without some parenthesis",
			Query: "(distance gt 100 or time lt 10000s) and date eq 2020-03-22",
			Result: bson.D{{"$and", []bson.D{
				{{"$or", []bson.D{
					{{"distance", bson.D{{"$gt", float32(100)}}}},
					{{"time", bson.D{{"$lt", time.Duration(10000 * time.Second)}}}},
				}}},
				{{"date", bson.D{{"$eq", tm}}}},
			}}},
		},
		{
			Name:  "correct without some parenthesis",
			Query: "(distance gt 100 or time lt 10000s) and (date eq 2020-03-22)",
			Result: bson.D{{"$and", []bson.D{
				{{"$or", []bson.D{
					{{"distance", bson.D{{"$gt", float32(100)}}}},
					{{"time", bson.D{{"$lt", time.Duration(10000 * time.Second)}}}},
				}}},
				{{"date", bson.D{{"$eq", tm}}}},
			}}},
		},
		{
			Name:  "missing parenthesis",
			Query: "(distance gt 100 or (time lt 10000s)) and (date eq 2020-03-22)",
			Result: bson.D{{"$and", []bson.D{
				{{"$or", []bson.D{
					{{"distance", bson.D{{"$gt", float32(100)}}}},
					{{"time", bson.D{{"$lt", time.Duration(10000 * time.Second)}}}},
				}}},
				{{"date", bson.D{{"$eq", tm}}}},
			}}},
		},
		{
			Name:  "without parenthesis",
			Query: "distance gt 100 or time lt 10000s and date eq 2020-03-22",
			Result: bson.D{{"$and", []bson.D{
				{{"$or", []bson.D{
					{{"distance", bson.D{{"$gt", float32(100)}}}},
					{{"time", bson.D{{"$lt", time.Duration(10000 * time.Second)}}}},
				}}},
				{{"date", bson.D{{"$eq", tm}}}},
			}}},
		},
		{
			Name:  "wrong place parenthesis: correct number or entities",
			Query: "distance gt (100 or time) lt 10000s and date eq 2020-03-22",
			Err:   ErrUnknownTerm,
		},
		{
			Name:  "wrong place parenthesis: with operator",
			Query: "distance gt (time gt 1000s)",
			Err:   ErrInvalidExpression,
		},
		{
			Name:  "term instead of expression: with correct second",
			Query: "distance and (time gt 1000s)",
			Err:   ErrInvalidExpression,
		},
		{
			Name:  "term instead of expression: two terms",
			Query: "time or time",
			Err:   ErrInvalidExpression,
		},
		{
			Name:  "wrong time value",
			Query: "time gt time",
			Err:   ErrInvalidValue,
		},
		{
			Name:  "wrong place parenthesis: around only number",
			Query: "distance gt (100) or time lt 10000s and date eq 2020-03-22",
			Err:   ErrUnknownTerm,
		},
		{
			Name:  "wrong place parenthesis: two operators",
			Query: "distance gt 100 (or time lt) 10000s and date eq 2020-03-22",
			Err:   ErrInvalidExpression,
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(tt *testing.T) {
			res, err := ParseTracking(tc.Query)
			if tc.Err != nil {
				assert.Equal(tt, tc.Err, err, "error is incorrect")

				return
			}
			assert.NoError(tt, err, "unexpected error")

			assert.Equal(tt, tc.Result, res, "result is incorrect")
		})
	}
}

func TestParseUsers(t *testing.T) {
	type TestCase struct {
		Name   string
		Query  string
		Result bson.D
		Err    error
	}

	tests := []TestCase{
		{
			Name:   "valid with parenthesis",
			Query:  "(email eq bahsjdbasd@gmail.com)",
			Result: bson.D{{"email", bson.D{{"$eq", "bahsjdbasd@gmail.com"}}}},
		},
		{
			Name:   "valid without parenthesis",
			Query:  "email eq bahsjdbasd@gmail.com",
			Result: bson.D{{"email", bson.D{{"$eq", "bahsjdbasd@gmail.com"}}}},
		},
		{
			Name:  "terms from tracking",
			Query: "date eq 2020-03-10",
			Err:   ErrUnknownTerm,
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(tt *testing.T) {
			res, err := ParseUsers(tc.Query)
			if tc.Err != nil {
				assert.Equal(tt, tc.Err, err, "error is incorrect")

				return
			}
			assert.NoError(tt, err, "unexpected error")

			assert.Equal(tt, tc.Result, res, "result is incorrect")
		})
	}
}
