// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package vtgate provides query routing rpc services
// for vttablets.
package vtgate

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/youtube/vitess/go/sqltypes"
	"github.com/youtube/vitess/go/vt/discovery"
	"github.com/youtube/vitess/go/vt/tabletserver/tabletconn"
	"github.com/youtube/vitess/go/vt/topo"
	"golang.org/x/net/context"

	querypb "github.com/youtube/vitess/go/vt/proto/query"
	topodatapb "github.com/youtube/vitess/go/vt/proto/topodata"
	vtgatepb "github.com/youtube/vitess/go/vt/proto/vtgate"
	vtrpcpb "github.com/youtube/vitess/go/vt/proto/vtrpc"
)

// This file uses the sandbox_test framework.

func TestResolverExecuteKeyspaceIds(t *testing.T) {
	testResolverGeneric(t, "TestResolverExecuteKeyspaceIds", func(hc discovery.HealthCheck) (*sqltypes.Result, error) {
		res := NewResolver(hc, topo.Server{}, new(sandboxTopo), "", "aa", 0, nil)
		return res.ExecuteKeyspaceIds(context.Background(),
			"query",
			nil,
			"TestResolverExecuteKeyspaceIds",
			[][]byte{{0x10}, {0x25}},
			topodatapb.TabletType_MASTER,
			nil,
			false)
	})
}

func TestResolverExecuteKeyRanges(t *testing.T) {
	testResolverGeneric(t, "TestResolverExecuteKeyRanges", func(hc discovery.HealthCheck) (*sqltypes.Result, error) {
		res := NewResolver(hc, topo.Server{}, new(sandboxTopo), "", "aa", 0, nil)
		return res.ExecuteKeyRanges(context.Background(),
			"query",
			nil,
			"TestResolverExecuteKeyRanges",
			[]*topodatapb.KeyRange{{Start: []byte{0x10}, End: []byte{0x25}}},
			topodatapb.TabletType_MASTER,
			nil,
			false)
	})
}

func TestResolverExecuteEntityIds(t *testing.T) {
	testResolverGeneric(t, "TestResolverExecuteEntityIds", func(hc discovery.HealthCheck) (*sqltypes.Result, error) {
		res := NewResolver(hc, topo.Server{}, new(sandboxTopo), "", "aa", 0, nil)
		return res.ExecuteEntityIds(context.Background(),
			"query",
			nil,
			"TestResolverExecuteEntityIds",
			"col",
			[]*vtgatepb.ExecuteEntityIdsRequest_EntityId{
				{
					Type:       sqltypes.Int64,
					Value:      []byte("0"),
					KeyspaceId: []byte{0x10},
				},
				{
					Type:       sqltypes.VarBinary,
					Value:      []byte("1"),
					KeyspaceId: []byte{0x25},
				},
			},
			topodatapb.TabletType_MASTER,
			nil,
			false)
	})
}

func TestResolverExecuteBatchKeyspaceIds(t *testing.T) {
	testResolverGeneric(t, "TestResolverExecuteBatchKeyspaceIds", func(hc discovery.HealthCheck) (*sqltypes.Result, error) {
		res := NewResolver(hc, topo.Server{}, new(sandboxTopo), "", "aa", 0, nil)
		qrs, err := res.ExecuteBatchKeyspaceIds(context.Background(),
			[]*vtgatepb.BoundKeyspaceIdQuery{{
				Query: &querypb.BoundQuery{
					Sql:           "query",
					BindVariables: nil,
				},
				Keyspace: "TestResolverExecuteBatchKeyspaceIds",
				KeyspaceIds: [][]byte{
					{0x10},
					{0x25},
				},
			}},
			topodatapb.TabletType_MASTER,
			false,
			nil)
		if err != nil {
			return nil, err
		}
		return &qrs[0], err
	})
}

func TestResolverStreamExecuteKeyspaceIds(t *testing.T) {
	keyspace := "TestResolverStreamExecuteKeyspaceIds"
	testResolverStreamGeneric(t, keyspace, func(hc discovery.HealthCheck) (*sqltypes.Result, error) {
		res := NewResolver(hc, topo.Server{}, new(sandboxTopo), "", "aa", 0, nil)
		qr := new(sqltypes.Result)
		err := res.StreamExecuteKeyspaceIds(context.Background(),
			"query",
			nil,
			keyspace,
			[][]byte{{0x10}, {0x15}},
			topodatapb.TabletType_MASTER,
			func(r *sqltypes.Result) error {
				appendResult(qr, r)
				return nil
			})
		return qr, err
	})
	testResolverStreamGeneric(t, keyspace, func(hc discovery.HealthCheck) (*sqltypes.Result, error) {
		res := NewResolver(hc, topo.Server{}, new(sandboxTopo), "", "aa", 0, nil)
		qr := new(sqltypes.Result)
		err := res.StreamExecuteKeyspaceIds(context.Background(),
			"query",
			nil,
			keyspace,
			[][]byte{{0x10}, {0x15}, {0x25}},
			topodatapb.TabletType_MASTER,
			func(r *sqltypes.Result) error {
				appendResult(qr, r)
				return nil
			})
		return qr, err
	})
}

func TestResolverStreamExecuteKeyRanges(t *testing.T) {
	keyspace := "TestResolverStreamExecuteKeyRanges"
	// streaming a single shard
	testResolverStreamGeneric(t, keyspace, func(hc discovery.HealthCheck) (*sqltypes.Result, error) {
		res := NewResolver(hc, topo.Server{}, new(sandboxTopo), "", "aa", 0, nil)
		qr := new(sqltypes.Result)
		err := res.StreamExecuteKeyRanges(context.Background(),
			"query",
			nil,
			keyspace,
			[]*topodatapb.KeyRange{{Start: []byte{0x10}, End: []byte{0x15}}},
			topodatapb.TabletType_MASTER,
			func(r *sqltypes.Result) error {
				appendResult(qr, r)
				return nil
			})
		return qr, err
	})
	// streaming multiple shards
	testResolverStreamGeneric(t, keyspace, func(hc discovery.HealthCheck) (*sqltypes.Result, error) {
		res := NewResolver(hc, topo.Server{}, new(sandboxTopo), "", "aa", 0, nil)
		qr := new(sqltypes.Result)
		err := res.StreamExecuteKeyRanges(context.Background(),
			"query",
			nil,
			keyspace,
			[]*topodatapb.KeyRange{{Start: []byte{0x10}, End: []byte{0x25}}},
			topodatapb.TabletType_MASTER,
			func(r *sqltypes.Result) error {
				appendResult(qr, r)
				return nil
			})
		return qr, err
	})
}

func testResolverGeneric(t *testing.T, name string, action func(hc discovery.HealthCheck) (*sqltypes.Result, error)) {
	// successful execute
	s := createSandbox(name)
	sbc0 := &sandboxConn{}
	sbc1 := &sandboxConn{}
	hc := newFakeHealthCheck()
	hc.addTestEndPoint("aa", "1.1.1.1", 1001, name, "-20", topodatapb.TabletType_MASTER, true, 1, nil, sbc0)
	hc.addTestEndPoint("aa", "1.1.1.1", 1002, name, "20-40", topodatapb.TabletType_MASTER, true, 1, nil, sbc1)

	_, err := action(hc)
	if err != nil {
		t.Errorf("want nil, got %v", err)
	}
	if execCount := sbc0.ExecCount.Get(); execCount != 1 {
		t.Errorf("want 1, got %v", execCount)
	}
	if execCount := sbc1.ExecCount.Get(); execCount != 1 {
		t.Errorf("want 1, got %v", execCount)
	}

	// non-retryable failure
	s.Reset()
	sbc0 = &sandboxConn{mustFailServer: 1}
	sbc1 = &sandboxConn{mustFailRetry: 1}
	hc.Reset()
	hc.addTestEndPoint("aa", "-20", 1, name, "-20", topodatapb.TabletType_MASTER, true, 1, nil, sbc0)
	hc.addTestEndPoint("aa", "20-40", 1, name, "20-40", topodatapb.TabletType_MASTER, true, 1, nil, sbc1)
	_, err = action(hc)
	want1 := fmt.Sprintf("shard, host: %s.-20.master, host:\"-20\" port_map:<key:\"vt\" value:1 > , error: err", name)
	want2 := fmt.Sprintf("shard, host: %s.20-40.master, host:\"20-40\" port_map:<key:\"vt\" value:1 > , retry: err", name)
	want := []string{want1, want2}
	sort.Strings(want)
	if err == nil {
		t.Errorf("want\n%v\ngot\n%v", want, err)
	} else {
		got := strings.Split(err.Error(), "\n")
		sort.Strings(got)
		if !reflect.DeepEqual(want, got) {
			t.Errorf("want\n%v\ngot\n%v", want, got)
		}
	}
	// Ensure that we tried only once
	if execCount := sbc0.ExecCount.Get(); execCount != 1 {
		t.Errorf("want 1, got %v", execCount)
	}
	if execCount := sbc1.ExecCount.Get(); execCount != 1 {
		t.Errorf("want 1, got %v", execCount)
	}
	// Ensure that we tried topo only once when mapping KeyspaceId/KeyRange to shards
	if s.SrvKeyspaceCounter != 1 {
		t.Errorf("want 1, got %v", s.SrvKeyspaceCounter)
	}

	// retryable failure, no sharding event
	s.Reset()
	sbc0 = &sandboxConn{mustFailRetry: 1}
	sbc1 = &sandboxConn{mustFailFatal: 1}
	hc.Reset()
	hc.addTestEndPoint("aa", "-20", 1, name, "-20", topodatapb.TabletType_MASTER, true, 1, nil, sbc0)
	hc.addTestEndPoint("aa", "20-40", 1, name, "20-40", topodatapb.TabletType_MASTER, true, 1, nil, sbc1)
	_, err = action(hc)
	want1 = fmt.Sprintf("shard, host: %s.-20.master, host:\"-20\" port_map:<key:\"vt\" value:1 > , retry: err", name)
	want2 = fmt.Sprintf("shard, host: %s.20-40.master, host:\"20-40\" port_map:<key:\"vt\" value:1 > , fatal: err", name)
	want = []string{want1, want2}
	sort.Strings(want)
	if err == nil {
		t.Errorf("want\n%v\ngot\n%v", want, err)
	} else {
		got := strings.Split(err.Error(), "\n")
		sort.Strings(got)
		if !reflect.DeepEqual(want, got) {
			t.Errorf("want\n%v\ngot\n%v", want, got)
		}
	}
	// Ensure that we tried only once.
	if execCount := sbc0.ExecCount.Get(); execCount != 1 {
		t.Errorf("want 1, got %v", execCount)
	}
	if execCount := sbc1.ExecCount.Get(); execCount != 1 {
		t.Errorf("want 1, got %v", execCount)
	}
	// Ensure that we tried topo only twice.
	if s.SrvKeyspaceCounter != 2 {
		t.Errorf("want 2, got %v", s.SrvKeyspaceCounter)
	}

	// no failure, initial vertical resharding
	s.Reset()
	addSandboxServedFrom(name, name+"ServedFrom0")
	sbc0 = &sandboxConn{}
	sbc1 = &sandboxConn{}
	hc.Reset()
	hc.addTestEndPoint("aa", "1.1.1.1", 1001, name, "-20", topodatapb.TabletType_MASTER, true, 1, nil, sbc0)
	hc.addTestEndPoint("aa", "1.1.1.1", 1002, name, "20-40", topodatapb.TabletType_MASTER, true, 1, nil, sbc1)
	s0 := createSandbox(name + "ServedFrom0") // make sure we have a fresh copy
	s0.ShardSpec = "-80-"
	sbc2 := &sandboxConn{}
	hc.addTestEndPoint("aa", "1.1.1.1", 1003, name+"ServedFrom0", "-80", topodatapb.TabletType_MASTER, true, 1, nil, sbc2)
	_, err = action(hc)
	if err != nil {
		t.Errorf("want nil, got %v", err)
	}
	// Ensure original keyspace is not used.
	if execCount := sbc0.ExecCount.Get(); execCount != 0 {
		t.Errorf("want 0, got %v", execCount)
	}
	if execCount := sbc1.ExecCount.Get(); execCount != 0 {
		t.Errorf("want 0, got %v", execCount)
	}
	// Ensure redirected keyspace is accessed once.
	if execCount := sbc2.ExecCount.Get(); execCount != 1 {
		t.Errorf("want 1, got %v", execCount)
	}
	// Ensure that we tried each keyspace only once.
	if s.SrvKeyspaceCounter != 1 {
		t.Errorf("want 1, got %v", s.SrvKeyspaceCounter)
	}
	if s0.SrvKeyspaceCounter != 1 {
		t.Errorf("want 1, got %v", s0.SrvKeyspaceCounter)
	}
	s0.Reset()

	// retryable failure, vertical resharding
	s.Reset()
	sbc0 = &sandboxConn{}
	sbc1 = &sandboxConn{mustFailFatal: 1}
	hc.Reset()
	hc.addTestEndPoint("aa", "1.1.1.1", 1001, name, "-20", topodatapb.TabletType_MASTER, true, 1, nil, sbc0)
	hc.addTestEndPoint("aa", "1.1.1.1", 1002, name, "20-40", topodatapb.TabletType_MASTER, true, 1, nil, sbc1)
	i := 0
	s.SrvKeyspaceCallback = func() {
		if i == 1 {
			addSandboxServedFrom(name, name+"ServedFrom")
			hc.Reset()
			hc.addTestEndPoint("aa", "1.1.1.1", 1001, name+"ServedFrom", "-20", topodatapb.TabletType_MASTER, true, 1, nil, sbc0)
			hc.addTestEndPoint("aa", "1.1.1.1", 1002, name+"ServedFrom", "20-40", topodatapb.TabletType_MASTER, true, 1, nil, sbc1)
		}
		i++
	}
	_, err = action(hc)
	if err != nil {
		t.Errorf("want nil, got %v", err)
	}
	// Ensure that we tried only twice.
	if execCount := sbc0.ExecCount.Get(); execCount != 2 {
		t.Errorf("want 2, got %v", execCount)
	}
	if execCount := sbc1.ExecCount.Get(); execCount != 2 {
		t.Errorf("want 2, got %v", execCount)
	}
	// Ensure that we tried topo only 3 times.
	if s.SrvKeyspaceCounter != 3 {
		t.Errorf("want 3, got %v", s.SrvKeyspaceCounter)
	}

	// retryable failure, horizontal resharding
	s.Reset()
	sbc0 = &sandboxConn{}
	sbc1 = &sandboxConn{mustFailRetry: 1}
	hc.Reset()
	hc.addTestEndPoint("aa", "1.1.1.1", 1001, name, "-20", topodatapb.TabletType_MASTER, true, 1, nil, sbc0)
	hc.addTestEndPoint("aa", "1.1.1.1", 1002, name, "20-40", topodatapb.TabletType_MASTER, true, 1, nil, sbc1)
	i = 0
	s.SrvKeyspaceCallback = func() {
		if i == 1 {
			s.ShardSpec = "-20-30-40-60-80-a0-c0-e0-"
			hc.Reset()
			hc.addTestEndPoint("aa", "1.1.1.1", 1001, name, "-20", topodatapb.TabletType_MASTER, true, 1, nil, sbc0)
			hc.addTestEndPoint("aa", "1.1.1.1", 1002, name, "20-30", topodatapb.TabletType_MASTER, true, 1, nil, sbc1)
		}
		i++
	}
	_, err = action(hc)
	if err != nil {
		t.Errorf("want nil, got %v", err)
	}
	// Ensure that we tried only twice.
	if execCount := sbc0.ExecCount.Get(); execCount != 2 {
		t.Errorf("want 2, got %v", execCount)
	}
	if execCount := sbc1.ExecCount.Get(); execCount != 2 {
		t.Errorf("want 2, got %v", execCount)
	}
	// Ensure that we tried topo only twice.
	if s.SrvKeyspaceCounter != 2 {
		t.Errorf("want 2, got %v", s.SrvKeyspaceCounter)
	}
}

func testResolverStreamGeneric(t *testing.T, name string, action func(hc discovery.HealthCheck) (*sqltypes.Result, error)) {
	// successful execute
	s := createSandbox(name)
	sbc0 := &sandboxConn{}
	sbc1 := &sandboxConn{}
	hc := newFakeHealthCheck()
	hc.addTestEndPoint("aa", "1.1.1.1", 1001, name, "-20", topodatapb.TabletType_MASTER, true, 1, nil, sbc0)
	hc.addTestEndPoint("aa", "1.1.1.1", 1002, name, "20-40", topodatapb.TabletType_MASTER, true, 1, nil, sbc1)
	_, err := action(hc)
	if err != nil {
		t.Errorf("want nil, got %v", err)
	}
	if execCount := sbc0.ExecCount.Get(); execCount != 1 {
		t.Errorf("want 1, got %v", execCount)
	}

	// failure
	s.Reset()
	sbc0 = &sandboxConn{mustFailRetry: 1}
	sbc1 = &sandboxConn{}
	hc.Reset()
	hc.addTestEndPoint("aa", "-20", 1, name, "-20", topodatapb.TabletType_MASTER, true, 1, nil, sbc0)
	hc.addTestEndPoint("aa", "20-40", 1, name, "20-40", topodatapb.TabletType_MASTER, true, 1, nil, sbc1)
	_, err = action(hc)
	want := fmt.Sprintf("shard, host: %s.-20.master, host:\"-20\" port_map:<key:\"vt\" value:1 > , retry: err", name)
	if err == nil || err.Error() != want {
		t.Errorf("want\n%s\ngot\n%v", want, err)
	}
	// Ensure that we tried only once.
	if execCount := sbc0.ExecCount.Get(); execCount != 1 {
		t.Errorf("want 1, got %v", execCount)
	}
	// Ensure that we tried topo only once
	if s.SrvKeyspaceCounter != 1 {
		t.Errorf("want 1, got %v", s.SrvKeyspaceCounter)
	}
}

func TestResolverInsertSqlClause(t *testing.T) {
	clause := "col in (:col1, :col2)"
	tests := [][]string{
		{
			"select a from table",
			"select a from table where " + clause},
		{
			"select a from table where id = 1",
			"select a from table where id = 1 and " + clause},
		{
			"select a from table group by a",
			"select a from table where " + clause + " group by a"},
		{
			"select a from table where id = 1 order by a limit 10",
			"select a from table where id = 1 and " + clause + " order by a limit 10"},
		{
			"select a from table where id = 1 for update",
			"select a from table where id = 1 and " + clause + " for update"},
	}
	for _, test := range tests {
		got := insertSQLClause(test[0], clause)
		if got != test[1] {
			t.Errorf("want '%v', got '%v'", test[1], got)
		}
	}
}

func TestResolverBuildEntityIds(t *testing.T) {
	shardMap := make(map[string][]interface{})
	shardMap["-20"] = []interface{}{"0", 1}
	shardMap["20-40"] = []interface{}{"2"}
	sql := "select a from table where id=:id"
	entityColName := "uid"
	bindVar := make(map[string]interface{})
	bindVar["id"] = 10
	shards, sqls, bindVars := buildEntityIds(shardMap, sql, entityColName, bindVar)
	wantShards := []string{"-20", "20-40"}
	wantSqls := map[string]string{
		"-20":   "select a from table where id=:id and uid in (:uid0, :uid1)",
		"20-40": "select a from table where id=:id and uid in (:uid0)",
	}
	wantBindVars := map[string]map[string]interface{}{
		"-20":   {"id": 10, "uid0": "0", "uid1": 1},
		"20-40": {"id": 10, "uid0": "2"},
	}
	sort.Strings(wantShards)
	sort.Strings(shards)
	if !reflect.DeepEqual(wantShards, shards) {
		t.Errorf("want %+v, got %+v", wantShards, shards)
	}
	if !reflect.DeepEqual(wantSqls, sqls) {
		t.Errorf("want %+v, got %+v", wantSqls, sqls)
	}
	if !reflect.DeepEqual(wantBindVars, bindVars) {
		t.Errorf("want %+v, got %+v", wantBindVars, bindVars)
	}
}

func TestResolverDmlOnMultipleKeyspaceIds(t *testing.T) {
	keyspace := "TestResolverDmlOnMultipleKeyspaceIds"
	createSandbox(keyspace)
	sbc0 := &sandboxConn{}
	sbc1 := &sandboxConn{}
	hc := newFakeHealthCheck()
	hc.addTestEndPoint("aa", "1.1.1.1", 1001, keyspace, "-20", topodatapb.TabletType_MASTER, true, 1, nil, sbc0)
	hc.addTestEndPoint("aa", "1.1.1.1", 1002, keyspace, "20-40", topodatapb.TabletType_MASTER, true, 1, nil, sbc1)

	res := NewResolver(hc, topo.Server{}, new(sandboxTopo), "", "aa", 0, nil)
	errStr := "DML should not span multiple keyspace_ids"
	_, err := res.ExecuteKeyspaceIds(context.Background(),
		"update table set a = b",
		nil,
		keyspace,
		[][]byte{{0x10}, {0x25}},
		topodatapb.TabletType_MASTER,
		nil,
		false)
	if err == nil {
		t.Errorf("want %v, got nil", errStr)
	}
}

func TestResolverExecBatchReresolve(t *testing.T) {
	keyspace := "TestResolverExecBatchReresolve"
	createSandbox(keyspace)
	sbc := &sandboxConn{mustFailRetry: 20}
	hc := newFakeHealthCheck()
	hc.addTestEndPoint("aa", "0", 1, keyspace, "0", topodatapb.TabletType_MASTER, true, 1, nil, sbc)

	res := NewResolver(hc, topo.Server{}, new(sandboxTopo), "", "aa", 0, nil)

	callcount := 0
	buildBatchRequest := func() (*scatterBatchRequest, error) {
		callcount++
		queries := []*vtgatepb.BoundShardQuery{{
			Query: &querypb.BoundQuery{
				Sql:           "query",
				BindVariables: nil,
			},
			Keyspace: keyspace,
			Shards:   []string{"0"},
		}}
		return boundShardQueriesToScatterBatchRequest(queries)
	}

	_, err := res.ExecuteBatch(context.Background(), topodatapb.TabletType_MASTER, false, nil, buildBatchRequest)
	want := "shard, host: TestResolverExecBatchReresolve.0.master, host:\"0\" port_map:<key:\"vt\" value:1 > , retry: err"
	if err == nil || err.Error() != want {
		t.Errorf("want %s, got %v", want, err)
	}
	// Ensure scatter tried a re-resolve
	if callcount != 2 {
		t.Errorf("want 2, got %v", callcount)
	}
	if count := sbc.AsTransactionCount.Get(); count != 0 {
		t.Errorf("want 0, got %v", count)
	}
}

func TestResolverExecBatchAsTransaction(t *testing.T) {
	keyspace := "TestResolverExecBatchAsTransaction"
	createSandbox(keyspace)
	sbc := &sandboxConn{mustFailRetry: 20}
	hc := newFakeHealthCheck()
	hc.addTestEndPoint("aa", "0", 1, keyspace, "0", topodatapb.TabletType_MASTER, true, 1, nil, sbc)

	res := NewResolver(hc, topo.Server{}, new(sandboxTopo), "", "aa", 0, nil)

	callcount := 0
	buildBatchRequest := func() (*scatterBatchRequest, error) {
		callcount++
		queries := []*vtgatepb.BoundShardQuery{{
			Query: &querypb.BoundQuery{
				Sql:           "query",
				BindVariables: nil,
			},
			Keyspace: keyspace,
			Shards:   []string{"0"},
		}}
		return boundShardQueriesToScatterBatchRequest(queries)
	}

	_, err := res.ExecuteBatch(context.Background(), topodatapb.TabletType_MASTER, true, nil, buildBatchRequest)
	want := "shard, host: TestResolverExecBatchAsTransaction.0.master, host:\"0\" port_map:<key:\"vt\" value:1 > , retry: err"
	if err == nil || err.Error() != want {
		t.Errorf("want %v, got %v", want, err)
	}
	// Ensure scatter did not re-resolve
	if callcount != 1 {
		t.Errorf("want 1, got %v", callcount)
	}
	if count := sbc.AsTransactionCount.Get(); count != 1 {
		t.Errorf("want 1, got %v", count)
		return
	}
}

func TestIsRetryableError(t *testing.T) {
	var connErrorTests = []struct {
		in      error
		outBool bool
	}{
		{fmt.Errorf("generic error"), false},
		{&ScatterConnError{Retryable: true}, true},
		{&ScatterConnError{Retryable: false}, false},
		{&ShardError{EndPointCode: vtrpcpb.ErrorCode_QUERY_NOT_SERVED}, true},
		{&ShardError{EndPointCode: vtrpcpb.ErrorCode_INTERNAL_ERROR}, false},
		// tabletconn.ServerError will not come directly here,
		// they'll be wrapped in ScatterConnError or ShardConnError.
		// So they can't be retried as is.
		{&tabletconn.ServerError{ServerCode: vtrpcpb.ErrorCode_QUERY_NOT_SERVED}, false},
		{&tabletconn.ServerError{ServerCode: vtrpcpb.ErrorCode_PERMISSION_DENIED}, false},
	}

	for _, tt := range connErrorTests {
		gotBool := isRetryableError(tt.in)
		if gotBool != tt.outBool {
			t.Errorf("isConnError(%v) => %v, want %v",
				tt.in, gotBool, tt.outBool)
		}
	}
}
