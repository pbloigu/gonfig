package series

type SeriesValue struct {
	Id         int
	CreatedAt  int
	RecordedAt *int
	Data       *string
}

type Series struct {
	Id                int
	LastValueTime     *int
	LastValueRecorded *int
	LastValue         *string
	Name              string
}
