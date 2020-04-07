package metrics

import (
    "errors"
    "fmt"
    "math/big"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/settings"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// RP metrics process
type RocketPoolMetricsProcess struct {
    p                       *services.Provider
    stakingDurationEnabled  map[string]prometheus.Gauge
    networkUtilisation      map[string]prometheus.Gauge
    rplRatio                map[string]prometheus.Gauge

}


// Staking duration metrics
type StakingDurationMetrics struct {
    DurationId string
    NetworkUtilisation float64
    RplRatio float64
}


// Start RP metrics process
func StartRocketPoolMetricsProcess(p *services.Provider) {

    // Initialise process / register metrics
    process := &RocketPoolMetricsProcess{
        p:                      p,
        stakingDurationEnabled: make(map[string]prometheus.Gauge),
        networkUtilisation:     make(map[string]prometheus.Gauge),
        rplRatio:               make(map[string]prometheus.Gauge),
    }

    // Start
    process.start()

}


// Start process
func (p *RocketPoolMetricsProcess) start() {

    // Update metrics on interval
    go (func() {
        p.update()
        updateMetricsTimer := time.NewTicker(updateMetricsInterval)
        for _ = range updateMetricsTimer.C {
            p.update()
        }
    })()

}


// Update metrics
func (p *RocketPoolMetricsProcess) update() {

    // Get minipool staking durations
    stakingDurations, err := settings.GetMinipoolStakingDurations(p.p.CM)
    if err != nil {
        p.p.Log.Println(err)
        return
    }
    stakingDurationCount := len(stakingDurations)

    // Get staking duration metrics
    stakingDurationMetricsChannels := make([]chan *StakingDurationMetrics, stakingDurationCount)
    errorChannel := make(chan error)
    for di := 0; di < stakingDurationCount; di++ {
        stakingDurationMetricsChannels[di] = make(chan *StakingDurationMetrics)
        go (func(di int) {
            if metrics, err := getStakingDurationMetrics(p.p.CM, stakingDurations[di].Id); err != nil {
                errorChannel <- err
            } else {
                stakingDurationMetricsChannels[di] <- metrics
            }
        })(di)
    }

    // Receive staking duration metrics
    stakingDurationMetrics := make([]*StakingDurationMetrics, stakingDurationCount)
    for di := 0; di < stakingDurationCount; di++ {
        select {
            case metrics := <-stakingDurationMetricsChannels[di]:
                stakingDurationMetrics[di] = metrics
            case err := <-errorChannel:
                p.p.Log.Println(err)
                return
        }
    }

    // Create & update metrics
    for di := 0; di < stakingDurationCount; di++ {
        duration := stakingDurations[di]
        metrics := stakingDurationMetrics[di]

        // Staking duration enabled
        if _, ok := p.stakingDurationEnabled[duration.Id]; !ok {
            p.stakingDurationEnabled[duration.Id] = promauto.NewGauge(prometheus.GaugeOpts{
                Namespace:  "smartnode",
                Subsystem:  "rocketpool",
                Name:       fmt.Sprintf("staking_enabled_%s", duration.Id),
                Help:       fmt.Sprintf("Whether the '%s' staking duration is enabled", duration.Id),
            })
        }
        if duration.Enabled {
            p.stakingDurationEnabled[duration.Id].Set(1)
        } else {
            p.stakingDurationEnabled[duration.Id].Set(0)
        }

        // Network utilisation
        if _, ok := p.networkUtilisation[duration.Id]; !ok {
            p.networkUtilisation[duration.Id] = promauto.NewGauge(prometheus.GaugeOpts{
                Namespace:  "smartnode",
                Subsystem:  "rocketpool",
                Name:       fmt.Sprintf("network_utilisation_%s", duration.Id),
                Help:       fmt.Sprintf("The network utilisation for the '%s' queue", duration.Id),
            })
        }
        p.networkUtilisation[duration.Id].Set(metrics.NetworkUtilisation)

        // RPL ratio
        if _, ok := p.rplRatio[duration.Id]; !ok {
            p.rplRatio[duration.Id] = promauto.NewGauge(prometheus.GaugeOpts{
                Namespace:  "smartnode",
                Subsystem:  "rocketpool",
                Name:       fmt.Sprintf("rpl_ratio_%s", duration.Id),
                Help:       fmt.Sprintf("The RPL:ETH ratio for the '%s' queue", duration.Id),
            })
        }
        p.rplRatio[duration.Id].Set(metrics.RplRatio)

    }

}


// Get staking duration metrics
func getStakingDurationMetrics(cm *rocketpool.ContractManager, durationId string) (*StakingDurationMetrics, error) {

    // Data channels
    networkUtilisationChannel := make(chan float64)
    rplRatioChannel := make(chan float64)
    errorChannel := make(chan error)

    // Get network utilisation
    go (func() {
        networkUtilisation := new(*big.Int)
        if err := cm.Contracts["rocketPool"].Call(nil, networkUtilisation, "getNetworkUtilisation", durationId); err != nil {
            errorChannel <- errors.New("Error retrieving network utilisation: " + err.Error())
        } else {
            networkUtilisationChannel <- eth.WeiToEth(*networkUtilisation)
        }
    })()

    // Get RPL ratio
    go (func() {
        rplRatioWei := new(*big.Int)
        if err := cm.Contracts["rocketNodeAPI"].Call(nil, rplRatioWei, "getRPLRatio", durationId); err != nil {
            errorChannel <- errors.New("Error retrieving required RPL amount: " + err.Error())
        } else {
            rplRatioChannel <- eth.WeiToEth(*rplRatioWei)
        }
    })()

    // Receive data
    var networkUtilisation float64
    var rplRatio float64
    for received := 0; received < 2; {
        select {
            case networkUtilisation = <-networkUtilisationChannel:
                received++
            case rplRatio = <-rplRatioChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return &StakingDurationMetrics{
        DurationId: durationId,
        NetworkUtilisation: networkUtilisation,
        RplRatio: rplRatio,
    }, nil

}

