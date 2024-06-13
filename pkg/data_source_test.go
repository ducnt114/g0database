package pkg

import "testing"

func TestCsvDataSource_FromFile(t *testing.T) {
	csvSource := NewDataSource(DataSourceTypeCsv)
	err := csvSource.FromFile("../resources/user.csv")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}
