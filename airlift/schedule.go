package airlift

import r "github.com/dancannon/gorethink"

// Schedule represents the schedule for a user.
type Schedule []Subject

// ScheduleFor returns the schedule for a user.
func ScheduleFor(user User) (Schedule, error) {
	results := Schedule{}
	err := getAll(r.
		Table("subjects").
		GetAll(r.Args(user.Schedule)).
		OrderBy(r.Asc("exam_time")),
		&results)
	if err != nil {
		return Schedule{}, err
	}
	return results, nil
}

// UpdateScheduleByMatch updates a user's schedule by using prefix matching
// based on data from app4.
func UpdateScheduleByMatch(username string, schedule []string) error {
	_, err := r.
		Table("subjects").
		Filter(func(row r.Term) interface{} {
			return r.Expr(schedule).Contains(func(class r.Term) interface{} {
				return class.Match(row.Field("match"))
			})
		}).
		OrderBy(r.Asc("exam_time")).
		Field("id").
		CoerceTo("array").
		Do(func(fullSchedule r.Term) interface{} {
			return r.Table("users").Get(username).Update(map[string]interface{}{
				"schedule":     fullSchedule,
				"raw_schedule": schedule,
			})
		}).
		RunWrite(session)
	return err
}
