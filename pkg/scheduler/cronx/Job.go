package cronx

type Job interface {
	CronExpr() string
	Run()
}
