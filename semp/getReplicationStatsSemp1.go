package semp

import (
	"encoding/xml"
	"errors"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Get DR replication statistics
func (e *Semp) GetReplicationStatsSemp1(ch chan<- PrometheusMetric) (ok float64, err error) {
	type Data struct {
		RPC struct {
			Show struct {
				Repl struct {
					Mate struct {
						Name string `xml:"router-name"`
					} `xml:"mate"`
					ConfigSync struct {
						Bridge struct {
							AdminState string `xml:"admin-state"`
							State      string `xml:"state"`
						} `xml:"bridge"`
					} `xml:"config-sync"`
					Stats struct {
						ActiveStats struct {
							MsgProcessing struct {
								SyncQ2Standby      float64 `xml:"sync-msgs-queued-to-standby"`
								SyncQ2StandbyAsync float64 `xml:"sync-msgs-queued-to-standby-as-async"`
								AsyncQ2Standby     float64 `xml:"async-msgs-queued-to-standby"`
								PromotedQ2Standby  float64 `xml:"promoted-msgs-queued-to-standby"`
								PrunedConsumed     float64 `xml:"pruned-locally-consumed-msgs"`
							} `xml:"message-processing"`
							SyncRepl struct {
								Trans2Ineligible float64 `xml:"transitions-to-ineligible"`
							} `xml:"sync-replication"`
							AckPropagation struct {
								TxMsgToStandby   float64 `xml:"msgs-tx-to-standby"`
								RxReqFromStandby float64 `xml:"rec-req-from-standby"`
							} `xml:"ack-propagation"`
						} `xml:"active-stats"`
						StandbyStats struct {
							MsgProcessing struct {
								RxMsgFromActive float64 `xml:"msgs-rx-from-active"`
							} `xml:"message-processing"`
							AckPropagation struct {
								RxAckPropMsgs float64 `xml:"ack-prop-msgs-rx"`
								TxReconReq    float64 `xml:"recon-req-tx"`
								RxOutOfSeq    float64 `xml:"out-of-seq-rx"`
							} `xml:"ack-propagation"`
							XaRepl struct {
								XaReq                float64 `xml:"transaction-request"`
								XaReqSuccess         float64 `xml:"transaction-request-success"`
								XaReqSuccessPrepare  float64 `xml:"transaction-request-success-prepare"`
								XaReqSuccessCommit   float64 `xml:"transaction-request-success-commit"`
								XaReqSuccessRollback float64 `xml:"transaction-request-success-rollback"`
								XaReqFail            float64 `xml:"transaction-request-fail"`
								XaReqFailPrepare     float64 `xml:"transaction-request-fail-prepare"`
								XaReqFailCommit      float64 `xml:"transaction-request-fail-commit"`
								XaReqFailRollback    float64 `xml:"transaction-request-fail-rollback"`
							} `xml:"transaction-replication"`
						} `xml:"standby-stats"`
					} `xml:"stats"`
				} `xml:"replication"`
			} `xml:"show"`
		} `xml:"rpc"`
		ExecuteResult struct {
			Result string `xml:"code,attr"`
		} `xml:"execute-result"`
	}

	command := "<rpc><show><replication><stats/></replication></show></rpc>"
	body, err := e.postHTTP(e.brokerURI+"/SEMP", "application/xml", command, "ReplicationStatsSemp1", 1)
	if err != nil {
		_ = level.Error(e.logger).Log("msg", "Can't scrape ReplicationStatsSemp1", "err", err, "broker", e.brokerURI)
		return 0, err
	}
	defer body.Close()
	decoder := xml.NewDecoder(body)
	var target Data
	err = decoder.Decode(&target)
	if err != nil {
		_ = level.Error(e.logger).Log("msg", "Can't decode Xml ReplicationStatsSemp1", "err", err, "broker", e.brokerURI)
		return 0, err
	}
	if target.ExecuteResult.Result != "ok" {
		_ = level.Error(e.logger).Log("msg", "unexpected result", "command", command, "result", target.ExecuteResult.Result, "broker", e.brokerURI)
		return 0, errors.New("unexpected result: see log")
	}

	replMateName := "" + target.RPC.Show.Repl.Mate.Name
	if replMateName != "" {
		replBridge := target.RPC.Show.Repl.ConfigSync.Bridge
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_bridge_admin_state"], prometheus.GaugeValue, encodeMetricMulti(replBridge.AdminState, []string{"Disabled", "Enabled", "-"}), replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_bridge_state"], prometheus.GaugeValue, encodeMetricMulti(replBridge.State, []string{"down", "up", "n/a"}), replMateName)
		//Active stats
		activeStats := target.RPC.Show.Repl.Stats.ActiveStats
		//Message processing
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_sync_msgs_queued_to_standby"], prometheus.GaugeValue, activeStats.MsgProcessing.SyncQ2Standby, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_sync_msgs_queued_to_standby_as_async"], prometheus.GaugeValue, activeStats.MsgProcessing.SyncQ2StandbyAsync, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_async_msgs_queued_to_standby"], prometheus.GaugeValue, activeStats.MsgProcessing.AsyncQ2Standby, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_promoted_msgs_queued_to_standby"], prometheus.GaugeValue, activeStats.MsgProcessing.PromotedQ2Standby, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_pruned_locally_consumed_msgs"], prometheus.GaugeValue, activeStats.MsgProcessing.PrunedConsumed, replMateName)
		//Sync replication
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_transitions_to_ineligible"], prometheus.GaugeValue, activeStats.SyncRepl.Trans2Ineligible, replMateName)
		//Ack propagation
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_msgs_tx_to_standby"], prometheus.GaugeValue, activeStats.AckPropagation.TxMsgToStandby, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_rec_req_from_standby"], prometheus.GaugeValue, activeStats.AckPropagation.RxReqFromStandby, replMateName)
		//Standby stats
		standbyStats := target.RPC.Show.Repl.Stats.StandbyStats
		//Message processing
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_msgs_rx_from_active"], prometheus.GaugeValue, standbyStats.MsgProcessing.RxMsgFromActive, replMateName)
		//Ack propagation
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_ack_prop_msgs_rx"], prometheus.GaugeValue, standbyStats.AckPropagation.RxAckPropMsgs, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_recon_req_tx"], prometheus.GaugeValue, standbyStats.AckPropagation.TxReconReq, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_out_of_seq_rx"], prometheus.GaugeValue, standbyStats.AckPropagation.RxOutOfSeq, replMateName)
		//Transaction replication
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_xa_req"], prometheus.GaugeValue, standbyStats.XaRepl.XaReq, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_xa_req_success"], prometheus.GaugeValue, standbyStats.XaRepl.XaReqSuccess, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_xa_req_success_prepare"], prometheus.GaugeValue, standbyStats.XaRepl.XaReqSuccessPrepare, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_xa_req_success_commit"], prometheus.GaugeValue, standbyStats.XaRepl.XaReqSuccessCommit, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_xa_req_success_rollback"], prometheus.GaugeValue, standbyStats.XaRepl.XaReqSuccessRollback, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_xa_req_fail"], prometheus.GaugeValue, standbyStats.XaRepl.XaReqFail, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_xa_req_fail_prepare"], prometheus.GaugeValue, standbyStats.XaRepl.XaReqFailPrepare, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_xa_req_fail_commit"], prometheus.GaugeValue, standbyStats.XaRepl.XaReqFailCommit, replMateName)
		ch <- e.NewMetric(MetricDesc["ReplicationStats"]["system_replication_xa_req_fail_rollback"], prometheus.GaugeValue, standbyStats.XaRepl.XaReqFailRollback, replMateName)
	}

	return 1, nil
}
