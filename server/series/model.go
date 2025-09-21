package series

type SeriesValue struct {
	CreatedAt int
	Data      *string
}

type Series struct {
	Id            int
	LastValueTime *int
	LastValue     *string
	Name          string
}
