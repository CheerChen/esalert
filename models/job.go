package models

type Job struct {
	Id        int64
	UserId    string `db:"user_id"`
	Name      string `db:"name"`
	Value     string `db:"value"`
	Status    int    `db:"status"`
	IsDeleted int    `db:"is_deleted"`
	//UpdatedAt time.Time `db:"updated_at"`
	//CreatedAt time.Time `db:"created_at"`
}

func GetJobById(id string) (job Job, err error) {
	err = db.Get(&job, "SELECT id,value,status,is_deleted FROM alert_job WHERE id=? LIMIT 1", id)
	if err != nil {
		return job, err
	}
	return job, nil
}

func GetJobs() (jobs []Job, err error) {
	err = db.Select(&jobs, "SELECT id,value,status,is_deleted FROM alert_job WHERE status=1 AND is_deleted=0")
	if err != nil {
		return jobs, err
	}
	return jobs, nil
}

func DelJobById(id string) (err error) {
	_, err = db.Exec("UPDATE alert_job SET is_deleted = 1 WHERE id=? LIMIT 1", id)
	if err != nil {
		return err
	}
	return nil
}
