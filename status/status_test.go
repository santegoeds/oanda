package status_test

import (
	"github.com/santegoeds/oanda/status"
	"gopkg.in/check.v1"
	"testing"
)

type TestSuite struct{}

var _ = check.Suite(&TestSuite{})

func Test(t *testing.T) { check.TestingT(t) }

func (ts *TestSuite) TestStatusApi(c *check.C) {
	services, err := status.Services()
	c.Assert(err, check.IsNil)
	c.Log(services)
	c.Assert(len(services) > 0, check.Equals, true)

	service, err := status.Service(services[0].Id)
	c.Assert(err, check.IsNil)
	c.Log(service)

	events, err := status.ServiceEvents(service.Id, nil, nil)
	c.Assert(err, check.IsNil)
	c.Log(events)

	currentEvent, err := status.CurrentServiceEvent(service.Id)
	c.Assert(err, check.IsNil)
	c.Log(currentEvent)
}

func (ts *TestSuite) TestServiceListApi(c *check.C) {
	serviceLists, err := status.ServiceLists()
	c.Assert(err, check.IsNil)
	c.Log(serviceLists)
	c.Assert(len(serviceLists) > 0, check.Equals, true)

	serviceList, err := status.ServiceList(serviceLists[0].Id)
	c.Assert(err, check.IsNil)
	c.Log(serviceList)
}

func (ts *TestSuite) TestServiceStatusApi(c *check.C) {
	statuses, err := status.ServiceStatuses()
	c.Assert(err, check.IsNil)
	c.Log(statuses)
	c.Assert(len(statuses) > 0, check.Equals, true)

	status, err := status.ServiceStatus(statuses[0].Id)
	c.Assert(err, check.IsNil)
	c.Log(status)
}
