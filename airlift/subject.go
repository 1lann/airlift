package airlift

import (
	"time"

	r "github.com/dancannon/gorethink"
)

// Subject represents a subject and its information.
type Subject struct {
	ID       string    `gorethink:"id"`
	Name     string    `gorethink:"name"`
	Match    string    `gorethink:"match"`
	ExamTime time.Time `gorethink:"exam_time"`
}

// FullSubject represents a subject along with notes and practice exams.
type FullSubject struct {
	Subject
	Notes  []Note  `gorethink:"notes"`
	Papers []Paper `gorethink:"papers"`
}

// AllSubjects returns all of the subjects in the database in alphabetical
// order.
func AllSubjects() ([]Subject, error) {
	var subjects []Subject
	err := getAll(r.Table("subjects").
		OrderBy(r.OrderByOpts{Index: r.Asc("name")}), &subjects)
	if err != nil {
		return nil, err
	}

	return subjects, nil
}

// GetSubject returns the basic information of a subject.
func GetSubject(id string) (Subject, error) {
	var subject Subject
	err := getOne(r.Table("subjects").Get(id), &subject)

	if err != nil {
		return Subject{}, err
	}

	return subject, nil
}

// GetFullSubject returns a subject along with its notes ordered by
// stars and papers ordered by publication year.
func GetFullSubject(id, username string) (FullSubject, error) {
	var subject FullSubject
	err := getOne(r.Table("subjects").Get(id).Default(map[string]string{}).
		Merge(map[string]interface{}{
			"notes": r.Table("notes").GetAllByIndex("subject", id).
				Merge(func(note r.Term) interface{} {
					return map[string]interface{}{
						"num_stars":   note.Field("stars").Count().Default(0),
						"has_starred": note.Field("stars").Contains(username),
					}
				}).OrderBy(r.Desc("num_stars")).CoerceTo("array"),
			"papers": r.Table("papers").GetAllByIndex("subject", id).
				OrderBy(r.Desc(rowFullPaperTitle)).Merge(func(paper r.Term) interface{} {
				return map[string]interface{}{
					"has_completed": paper.Field("completed").Contains(username),
				}
			}).CoerceTo("array"),
		}), &subject)

	if err != nil {
		return FullSubject{}, err
	}

	return subject, nil
}
