package opsgenie

import (
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func TestStart(t *testing.T) {
	TestingT(t)
}

type DailyMessengerTestSuite struct{}

var _ = Suite(&DailyMessengerTestSuite{})

func (suite *DailyMessengerTestSuite) TestJoinDutiesByUserName(c *C) {
	usersOnDuty := []UserOnDuty{
		{
			Name:  "User1",
			Start: time.Date(2016, time.May, 17, 0, 0, 0, 0, time.Local),
			End:   time.Date(2016, time.May, 17, 9, 0, 0, 0, time.Local),
		},
		{
			Name:  "User2",
			Start: time.Date(2016, time.May, 17, 9, 0, 0, 0, time.Local),
			End:   time.Date(2016, time.May, 17, 18, 0, 0, 0, time.Local),
		},
		{
			Name:  "User2",
			Start: time.Date(2016, time.May, 17, 18, 0, 0, 0, time.Local),
			End:   time.Date(2016, time.May, 18, 9, 0, 0, 0, time.Local),
		},
		{
			Name:  "User3",
			Start: time.Date(2016, time.May, 18, 9, 0, 0, 0, time.Local),
			End:   time.Date(2016, time.May, 18, 18, 0, 0, 0, time.Local),
		},
		{
			Name:  "User3",
			Start: time.Date(2016, time.May, 18, 18, 0, 0, 0, time.Local),
			End:   time.Date(2016, time.May, 19, 0, 0, 0, 0, time.Local),
		},
	}

	result := JoinDutiesByUserName(JoinDuties(usersOnDuty))

	c.Assert(len(result), Equals, 3)
	duties, found := result["User1"]
	c.Assert(found, Equals, true)
	c.Assert(len(duties), Equals, 1)
	c.Assert(duties[0].Name, Equals, "User1")
	c.Assert(duties[0].Start, Equals, time.Date(2016, time.May, 17, 0, 0, 0, 0, time.Local))
	c.Assert(duties[0].End, Equals, time.Date(2016, time.May, 17, 9, 0, 0, 0, time.Local))
	duties, found = result["User2"]
	c.Assert(found, Equals, true)
	c.Assert(len(duties), Equals, 1)
	c.Assert(duties[0].Name, Equals, "User2")
	c.Assert(duties[0].Start, Equals, time.Date(2016, time.May, 17, 9, 0, 0, 0, time.Local))
	c.Assert(duties[0].End, Equals, time.Date(2016, time.May, 18, 9, 0, 0, 0, time.Local))
	duties, found = result["User3"]
	c.Assert(found, Equals, true)
	c.Assert(len(duties), Equals, 1)
	c.Assert(duties[0].Name, Equals, "User3")
	c.Assert(duties[0].Start, Equals, time.Date(2016, time.May, 18, 9, 0, 0, 0, time.Local))
	c.Assert(duties[0].End, Equals, time.Date(2016, time.May, 19, 0, 0, 0, 0, time.Local))
}

func (suite *DailyMessengerTestSuite) TestJoinDutiesByUserName_Overlapped(c *C) {
	usersOnDuty := []UserOnDuty{
		{
			Name:  "User1",
			Start: time.Date(2016, time.May, 17, 0, 0, 0, 0, time.Local),
			End:   time.Date(2016, time.May, 17, 9, 0, 0, 0, time.Local),
		},
		{
			Name:  "User2",
			Start: time.Date(2016, time.May, 17, 9, 0, 0, 0, time.Local),
			End:   time.Date(2016, time.May, 17, 18, 0, 0, 0, time.Local),
		},
		{
			Name:  "User2",
			Start: time.Date(2016, time.May, 17, 18, 0, 0, 0, time.Local),
			End:   time.Date(2016, time.May, 18, 9, 0, 0, 0, time.Local),
		},
		{
			Name:  "User1",
			Start: time.Date(2016, time.May, 18, 9, 0, 0, 0, time.Local),
			End:   time.Date(2016, time.May, 18, 18, 0, 0, 0, time.Local),
		},
		{
			Name:  "User1",
			Start: time.Date(2016, time.May, 18, 18, 0, 0, 0, time.Local),
			End:   time.Date(2016, time.May, 19, 0, 0, 0, 0, time.Local),
		},
	}

	result := JoinDutiesByUserName(JoinDuties(usersOnDuty))

	c.Assert(len(result), Equals, 2)
	duties, found := result["User1"]
	c.Assert(found, Equals, true)
	c.Assert(len(duties), Equals, 2)
	c.Assert(duties[0].Name, Equals, "User1")
	c.Assert(duties[0].Start, Equals, time.Date(2016, time.May, 17, 0, 0, 0, 0, time.Local))
	c.Assert(duties[0].End, Equals, time.Date(2016, time.May, 17, 9, 0, 0, 0, time.Local))
	c.Assert(duties[1].Name, Equals, "User1")
	c.Assert(duties[1].Start, Equals, time.Date(2016, time.May, 18, 9, 0, 0, 0, time.Local))
	c.Assert(duties[1].End, Equals, time.Date(2016, time.May, 19, 0, 0, 0, 0, time.Local))
	duties, found = result["User2"]
	c.Assert(found, Equals, true)
	c.Assert(len(duties), Equals, 1)
	c.Assert(duties[0].Name, Equals, "User2")
	c.Assert(duties[0].Start, Equals, time.Date(2016, time.May, 17, 9, 0, 0, 0, time.Local))
	c.Assert(duties[0].End, Equals, time.Date(2016, time.May, 18, 9, 0, 0, 0, time.Local))
}
