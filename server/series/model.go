package series

type SeriesValue struct {
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
