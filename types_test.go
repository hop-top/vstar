// SPDX-License-Identifier: Apache-2.0

package vstar

import "testing"

func TestCompType_WireStrings(t *testing.T) {
	cases := []struct {
		got  CompType
		want string
	}{
		{CompCalendar, "VCALENDAR"},
		{CompTodo, "VTODO"},
		{CompJournal, "VJOURNAL"},
		{CompEvent, "VEVENT"},
		{CompFreeBusy, "VFREEBUSY"},
		{CompTimezone, "VTIMEZONE"},
		{CompAlarm, "VALARM"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if string(tc.got) != tc.want {
				t.Errorf("CompType = %q, want %q", string(tc.got), tc.want)
			}
		})
	}
}

func TestKind_WireStrings(t *testing.T) {
	cases := []struct {
		got  Kind
		want string
	}{
		{KindIndividual, "individual"},
		{KindOrg, "org"},
		{KindGroup, "group"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if string(tc.got) != tc.want {
				t.Errorf("Kind = %q, want %q", string(tc.got), tc.want)
			}
		})
	}
}

func TestTodoStatus_WireStrings(t *testing.T) {
	cases := []struct {
		got  TodoStatus
		want string
	}{
		{TodoNeedsAction, "NEEDS-ACTION"},
		{TodoInProcess, "IN-PROCESS"},
		{TodoCompleted, "COMPLETED"},
		{TodoCancelled, "CANCELLED"}, //nolint:misspell // RFC 5545 wire string.
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if string(tc.got) != tc.want {
				t.Errorf("TodoStatus = %q, want %q", string(tc.got), tc.want)
			}
		})
	}
}
