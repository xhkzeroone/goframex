package cronx

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"log"
	"reflect"
)

var cronParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

type Cron struct {
	*cron.Cron
}

func New() *Cron {
	return &Cron{
		Cron: cron.New(cron.WithSeconds()),
	}
}

func resolveCronExpr(expr string) (string, error) {
	if isValidCronExpr(expr) {
		return expr, nil
	}
	expr2 := viper.GetString(expr)
	if isValidCronExpr(expr2) {
		return expr2, nil
	}
	return "", fmt.Errorf("invalid cronx expression '%s' (and not found in config)", expr)
}

func (c *Cron) AddJob(cronExpr string, jobFunc func()) error {
	expr, err := resolveCronExpr(cronExpr)
	if err != nil {
		return err
	}
	if _, err := c.AddFunc(expr, jobFunc); err != nil {
		return fmt.Errorf("failed to add cronx job: %v", err)
	}
	return nil
}

func (c *Cron) AddJobs(jobs ...Job) {
	if len(jobs) == 0 {
		return
	}
	for _, job := range jobs {
		jobName := reflect.TypeOf(job).String()
		expr, err := resolveCronExpr(job.CronExpr())
		if err != nil {
			log.Printf("invalid cronx expr for job %s: %v", jobName, err)
			continue
		}
		if err := c.AddJob(expr, job.Run); err != nil {
			log.Printf("failed to add cronx job %s: %v", jobName, err)
			continue
		}
		log.Printf("registered cronx job %s với biểu thức [%s]", jobName, expr)
	}
}

func isValidCronExpr(expr string) bool {
	_, err := cronParser.Parse(expr)
	return err == nil
}
