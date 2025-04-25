package measurements

type MeasurementValue struct {
	CreatedAt int
	Data      *string
}

type Measurement struct {
	Id            int
	LastValueTime *int
	LastValue     *string
	Name          string
}
