package types

var (
	COND_EQ      = QueryConditionType{"eq", 1}
	COND_NE      = QueryConditionType{"ne", 1}
	COND_NULL    = QueryConditionType{"null", 0}
	COND_NOTNULL = QueryConditionType{"notnull", 0}
	COND_IN      = QueryConditionType{"in", -1}
	COND_NOTIN   = QueryConditionType{"notin", -1}
	COND_OR      = QueryConditionType{"or", 1}
	COND_AND     = QueryConditionType{"and", 1}

	mods = map[string]QueryConditionType{
		COND_EQ.Name:      COND_EQ,
		COND_NE.Name:      COND_NE,
		COND_NULL.Name:    COND_NULL,
		COND_NOTNULL.Name: COND_NOTNULL,
		COND_IN.Name:      COND_IN,
		COND_NOTIN.Name:   COND_NOTIN,
		COND_OR.Name:      COND_OR,
		COND_AND.Name:     COND_AND,
	}
)

type QueryConditionType struct {
	Name string
	Args int
}

type QueryCondition struct {
	Field         string
	Values        []interface{}
	conditionType QueryConditionType
	left, right   *QueryCondition
}

func (q *QueryCondition) ToCondition() Condition {
	cond := Condition{
		Modifier: q.conditionType.Name,
	}
	if q.conditionType.Args == 1 && len(q.Values) > 0 {
		cond.Value = q.Values[0]
	} else if q.conditionType.Args == -1 {
		cond.Value = q.Values
	}

	return cond
}

func ValidMod(mod string) bool {
	_, ok := mods[mod]
	return ok
}

func NewConditionFromString(field, mod string, values ...interface{}) *QueryCondition {
	return &QueryCondition{
		Field:         field,
		Values:        values,
		conditionType: mods[mod],
	}
}

func NewCondition(mod QueryConditionType, values ...interface{}) *QueryCondition {
	return &QueryCondition{
		Values:        values,
		conditionType: mod,
	}
}

func NE(value interface{}) *QueryCondition {
	return NewCondition(COND_NE, value)
}

func EQ(value interface{}) *QueryCondition {
	return NewCondition(COND_EQ, value)
}

func NULL(value interface{}) *QueryCondition {
	return NewCondition(COND_NULL)
}

func NOTNULL(value interface{}) *QueryCondition {
	return NewCondition(COND_NOTNULL)
}

func IN(values ...interface{}) *QueryCondition {
	return NewCondition(COND_IN, values...)
}

func NOTIN(values ...interface{}) *QueryCondition {
	return NewCondition(COND_NOTIN, values...)
}

func (c *QueryCondition) AND(right *QueryCondition) *QueryCondition {
	return &QueryCondition{
		conditionType: COND_AND,
		left:          c,
		right:         right,
	}
}

func (c *QueryCondition) OR(right *QueryCondition) *QueryCondition {
	return &QueryCondition{
		conditionType: COND_OR,
		left:          c,
		right:         right,
	}
}
