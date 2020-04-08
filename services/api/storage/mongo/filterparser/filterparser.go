package filterparser

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/boodyvo/jogging-api/lib"

	"gopkg.in/mgo.v2/bson"
)

// TODO(boodyvo): Fill all terms or make some reflect/codegen from pb to construct it automatically.

var (
	emailRegexp = regexp.MustCompile(`^[A-z0-9!#$%&'*+\/=?^_` + "`" + `{|}~-]+(?:\.[A-z0-9!#$%&'*+\/=?^_` + "`" + `{|}~-]+)*@(?:[A-z0-9](?:[A-z0-9-]*[A-z0-9])?\.)+[A-z0-9](?:[A-z0-9-]*[A-z0-9])?$`)
)

type Operation struct {
	Name string
	Calc func(ex1, ex2 bson.D) bson.D
}

type Operator struct {
	Name string
	Calc func(variable, value string) bson.D
}

type Checker func(string) (interface{}, error)

func ToTime(value string) (interface{}, error) {
	return time.Parse(lib.DateFormat, value)
}
func ToDuration(value string) (interface{}, error) {
	return time.ParseDuration(value)
}
func ToInt64(value string) (interface{}, error) {
	return strconv.ParseInt(value, 10, 64)
}
func ToInt32(value string) (interface{}, error) {
	res, err := strconv.ParseInt(value, 10, 32)

	return int32(res), err
}
func ToFloat64(value string) (interface{}, error) {
	return strconv.ParseFloat(value, 64)
}
func ToFloat32(value string) (interface{}, error) {
	res, err := strconv.ParseFloat(value, 32)

	return float32(res), err
}
func ToString(value string) (interface{}, error) {
	for _, charVariable := range value {
		if (charVariable < 'a' || charVariable > 'z') &&
			(charVariable < 'A' || charVariable > 'Z') {
			return nil, ErrInvalidValue
		}
	}

	return value, nil
}
func ToEmail(value string) (interface{}, error) {
	if !emailRegexp.MatchString(value) {
		return nil, ErrInvalidValue
	}

	return value, nil
}

func and(ex1, ex2 bson.D) bson.D {
	return bson.D{{"$and", []bson.D{ex1, ex2}}}
}
func or(ex1, ex2 bson.D) bson.D {
	return bson.D{{"$or", []bson.D{ex1, ex2}}}
}
func gt(variable string, value interface{}) bson.D {
	return bson.D{{variable, bson.D{{"$gt", value}}}}
}
func lt(variable string, value interface{}) bson.D {
	return bson.D{{variable, bson.D{{"$lt", value}}}}
}
func eq(variable string, value interface{}) bson.D {
	return bson.D{{variable, bson.D{{"$eq", value}}}}
}
func ne(variable string, value interface{}) bson.D {
	return bson.D{{variable, bson.D{{"$ne", value}}}}
}

var (
	operators = map[string]func(variable string, value interface{}) bson.D{
		"eq": eq,
		"ne": ne,
		"gt": gt,
		"lt": lt,
	}
	operations = map[string]func(ex1, ex2 bson.D) bson.D{
		"and": and,
		"or":  or,
	}
	termsTracking = map[string]Checker{
		"date":                    ToTime,
		"time":                    ToDuration,
		"distance":                ToFloat32,
		"location.longitude":      ToFloat64,
		"location.latitude":       ToFloat64,
		"weather.temperature":     ToFloat32,
		"weather.temperature_min": ToFloat32,
		"weather.temperature_max": ToFloat32,
		"weather.snowdepth":       ToFloat32,
		"weather.winddirection":   ToFloat32,
		"weather.windspeed":       ToFloat32,
		"weather.pressure":        ToFloat32,
	}
	termsUser = map[string]Checker{"email": ToEmail}
)

func isTerm(value string, terms map[string]Checker) bool {
	_, ok := terms[value]

	return ok
}

func isOperation(value string) bool {
	_, ok := operations[value]

	return ok
}

func isOperator(value string) bool {
	_, ok := operators[value]

	return ok
}

func parseParenthesis(query string) (map[int]int, error) {
	open := make([]int, 0, 1+len(query)/4)
	closed := make(map[int]int)
	for i, symb := range query {
		switch string(symb) {
		case "(":
			open = append(open, i)
		case ")":
			if len(open) == 0 {
				return nil, ErrIncorrectParenthesisQuery
			}

			closed[open[len(open)-1]] = i
			open = open[:len(open)-1]
		}
	}
	if len(open) > 0 {
		return nil, ErrIncorrectParenthesisQuery
	}

	return closed, nil
}

func calcExpressions(exp1 interface{}, o string, exp2 interface{}, terms map[string]Checker) (bson.D, error) {
	if isOperation(o) {
		expression1, ok := exp1.(bson.D)
		if !ok {
			return nil, ErrInvalidExpression
		}
		expression2, ok := exp2.(bson.D)
		if !ok {
			return nil, ErrInvalidExpression
		}

		return operations[o](expression1, expression2), nil
	}

	if isOperator(o) {
		variable, ok := exp1.(string)
		if !ok {
			return nil, ErrInvalidExpression
		}
		value, ok := exp2.(string)
		if !ok {
			return nil, ErrInvalidExpression
		}
		checker, ok := terms[variable]
		if !ok {
			return nil, ErrInvalidExpression
		}
		v, err := checker(value)
		if err != nil {
			return nil, ErrInvalidValue
		}

		return operators[o](variable, v), nil
	}

	return nil, ErrUnknownOperand
}

func extract(from, to int, query string, terms map[string]Checker, parenthesis map[int]int) (bson.D, error) {
	expressions := make([]interface{}, 0, 5)

	values := strings.Split(query, " ()")
	if len(values) == 0 {
		return nil, ErrEmptyQuery
	}

	cur := from
	for cur < to {
		if string(query[cur]) == " " {
			cur++
			continue
		}
		if string(query[cur]) == "(" {
			// expression should be first or third
			if len(expressions) != 0 && len(expressions) != 2 {
				return nil, ErrInvalidExpression
			}
			expression, err := extract(cur+1, parenthesis[cur], query, terms, parenthesis)
			if err != nil {
				return nil, err
			}

			expressions = append(expressions, expression)
			if len(expressions) == 3 {
				exp, err := calcExpressions(expressions[0], expressions[1].(string), expressions[2], terms)
				if err != nil {
					return nil, err
				}

				expressions = []interface{}{exp}
			}
			cur = parenthesis[cur] + 1

			continue
		}
		if string(query[cur]) == ")" {
			return expressions[0].(bson.D), nil
		}

		str := ""
		for cur < to && string(query[cur]) != " " &&
			string(query[cur]) != ")" && string(query[cur]) != "(" {
			str += string(query[cur])
			cur++
		}

		switch len(expressions) {
		case 0:
			if !isTerm(str, terms) {
				return nil, ErrUnknownTerm
			}

			expressions = append(expressions, str)

			cur++

			break
		case 1:
			if !isOperator(str) && !isOperation(str) {
				return nil, ErrUnknownOperand
			}

			expressions = append(expressions, str)
			cur++

			break
		case 2:
			expressions = append(expressions, str)
			if isOperator(expressions[1].(string)) {
				exp, err := calcExpressions(expressions[0], expressions[1].(string), expressions[2], terms)
				if err != nil {
					return nil, err
				}
				expressions = []interface{}{exp}
				cur++
				break
			}

			if !isTerm(str, terms) {
				return nil, ErrUnknownTerm
			}

			cur++
			break
		case 3:
			if !isOperator(str) {
				return nil, ErrUnknownOperand
			}
			expressions = append(expressions, str)

			cur++
			break
		case 4:
			expressions = append(expressions, str)
			exp, err := calcExpressions(expressions[2], expressions[3].(string), expressions[4], terms)
			if err != nil {
				return nil, err
			}
			exp, err = calcExpressions(expressions[0], expressions[1].(string), exp, terms)
			if err != nil {
				return nil, err
			}

			expressions = []interface{}{exp}

			cur++
			break
		default:
			return nil, ErrInvalidExpression
		}
	}

	if len(expressions) != 1 {
		return nil, ErrInvalidExpression
	}

	return expressions[0].(bson.D), nil
}

func ParseTracking(query string) (bson.D, error) {
	query = strings.TrimSpace(query)
	parenthesis, err := parseParenthesis(query)
	if err != nil {
		return nil, err
	}

	return extract(0, len(query), query, termsTracking, parenthesis)
}

func ParseUsers(query string) (bson.D, error) {
	query = strings.TrimSpace(query)
	parenthesis, err := parseParenthesis(query)
	if err != nil {
		return nil, err
	}

	return extract(0, len(query), query, termsUser, parenthesis)
}
