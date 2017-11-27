package fullname_parser

import (
	"reflect"
	"testing"
)

func TestParseFullname(t *testing.T) {
	tests := []struct {
		name           string
		fullname       string
		wantParsedName ParsedName
	}{
		{"base test", "Juan Xavier", ParsedName{First: "Juan", Last: "Xavier", rawName: "Juan Xavier", nameParts: []string{}, nameCommas: []bool{}}},
		{"title test", "Dr. Juan Xavier", ParsedName{Title: "Dr.", First: "Juan", Last: "Xavier", rawName: "Dr. Juan Xavier", nameParts: []string{}, nameCommas: []bool{}}},
		{"nick test", "Dr. Juan Xavier (Doc Vega)", ParsedName{Title: "Dr.", First: "Juan", Last: "Xavier", Nick: "Doc Vega", rawName: "Dr. Juan Xavier", nameParts: []string{}, nameCommas: []bool{}}},
		{"middle test", "Juan Q. Xavier", ParsedName{First: "Juan", Middle: "Q.", Last: "Xavier", rawName: "Juan Q. Xavier", nameParts: []string{}, nameCommas: []bool{}}},
		{"suffixes test", "Juan Xavier III (Doc Vega), Jr.", ParsedName{First: "Juan", Last: "Xavier", Nick: "Doc Vega", Suffix: "III, Jr.", rawName: "Juan Xavier III, Jr.", nameParts: []string{}, nameCommas: []bool{}}},
		{"full test", "de la Vega, Dr. Juan et Glova (Doc Vega) Q. Xavier III, Jr., Genius", ParsedName{Title: "Dr.", First: "Juan et Glova", Middle: "Q. Xavier", Last: "de la Vega", Nick: "Doc Vega", Suffix: "III, Jr., Genius", rawName: "de la Vega, Dr. Juan et Glova Q. Xavier III, Jr., Genius", nameParts: []string{}, nameCommas: []bool{}}},
		{"just last name", "Cotter", ParsedName{Last: "Cotter", rawName: "Cotter", nameParts: []string{}, nameCommas: []bool{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotParsedName := ParseFullname(tt.fullname); !reflect.DeepEqual(gotParsedName, tt.wantParsedName) {
				t.Errorf("ParseFullname() = %v, want %v", gotParsedName, tt.wantParsedName)
			}
		})
	}
}
