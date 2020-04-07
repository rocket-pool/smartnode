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
    networkEthCapacity      map[string]prometheus.Gauge
    networkEthAssigned      map[string]prometheus.Gauge
    networkUtilisation      map[string]prometheus.Gauge
    rplRatio                map[string]prometheus.Gauge
    queueBalance            map[string]prometheus.Gauge

}


// Staking duration metrics
type StakingDurationMetrics struct {
    DurationId string
    NetworkEthCapacity float64
    NetworkEthAssigned float64
    NetworkUtilisation float64
    RplRatio float64
    QueueBalance float64
}


// Start RP metrics process
func StartRocketPoolMetricsProcess(p *services.Provider) {

    // Initialise process / register metrics
    process := &RocketPoolMetricsProcess{
        p:                      p,
        stakingDurationEnabled: make(map[string]prometheus.Gauge),
        networkEthCapacity:     make(map[string]prometheus.Gauge),
        networkEthAssigned:     make(map[string]prometheus.Gauge),
        networkUtilisation:     make(map[string]prometheus.Gauge),
        rplRatio:               make(map[string]prometheus.Gauge),
        queueBalance:           make(map[string]prometheus.Gauge),
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

        // Network ETH capacity
        if _, ok := p.networkEthCapacity[duration.Id]; !ok {
            p.networkEthCapacity[duration.Id] = promauto.NewGauge(prometheus.GaugeOpts{
                Namespace:  "smartnode",
                Subsystem:  "rocketpool",
                Name:       fmt.Sprintf("network_eth_capacity_%s", duration.Id),
                Help:       fmt.Sprintf("The network ETH capacity for the '%s' queue", duration.Id),
            })
        }
        p.networkEthCapacity[duration.Id].Set(metrics.NetworkEthCapacity)

        // Network ETH assigned
        if _, ok := p.networkEthAssigned[duration.Id]; !ok {
            p.networkEthAssigned[duration.Id] = promauto.NewGauge(prometheus.GaugeOpts{
                Namespace:  "smartnode",
                Subsystem:  "rocketpool",
                Name:       fmt.Sprintf("network_eth_assigned_%s", duration.Id),
                Help:       fmt.Sprintf("The network ETH assigned for the '%s' queue", duration.Id),
            })
        }
        p.networkEthAssigned[duration.Id].Set(metrics.NetworkEthAssigned)

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

        // Queue balance
        if _, ok := p.queueBalance[duration.Id]; !ok {
            p.queueBalance[duration.Id] = promauto.NewGauge(prometheus.GaugeOpts{
                Namespace:  "smartnode",
                Subsystem:  "rocketpool",
                Name:       fmt.Sprintf("queue_balance_%s", duration.Id),
                Help:       fmt.Sprintf("The current ETH balance of the '%s' deposit queue", duration.Id),
            })
        }
        p.queueBalance[duration.Id].Set(metrics.QueueBalance)

    }

}


// Get staking duration metrics
func getStakingDurationMetrics(cm *rocketpool.ContractManager, durationId string) (*StakingDurationMetrics, error) {

    // Data channels
    networkEthCapacityChannel := make(chan float64)
    networkEthAssignedChannel := make(chan float64)
    networkUtilisationChannel := make(chan float64)
    rplRatioChannel := make(chan float64)
    queueBalanceChannel := make(chan float64)
    errorChannel := make(chan error)

    // Get network ETH capacity
    go (func() {
        networkEthCapacity := new(*big.Int)
        if err := cm.Contracts["rocketPool"].Call(nil, networkEthCapacity, "getTotalEther", "capacity", durationId); err != nil {
            errorChannel <- errors.New("Error retrieving network ETH capacity: " + err.Error())
        } else {
            networkEthCapacityChannel <- eth.WeiToEth(*networkEthCapacity)
        }
    })()

    // Get network ETH assigned
    go (func() {
        networkEthAssigned := new(*big.Int)
        if err := cm.Contracts["rocketPool"].Call(nil, networkEthAssigned, "getTotalEther", "assigned", durationId); err != nil {
            errorChannel <- errors.New("Error retrieving network ETH assigned: " + err.Error())
        } else {
            networkEthAssignedChannel <- eth.WeiToEth(*networkEthAssigned)
        }
    })()

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

    // Get queue balance
    go (func() {
        queueBalanceWei := new(*big.Int)
        if err := cm.Contracts["rocketDepositQueue"].Call(nil, queueBalanceWei, "getBalance", durationId); err != nil {
            errorChannel <- errors.New("Error retrieving deposit queue balance: " + err.Error())
        } else {
            queueBalanceChannel <- eth.WeiToEth(*queueBalanceWei)
        }
    })()

    // Receive data
    var networkEthCapacity float64
    var networkEthAssigned float64
    var networkUtilisation float64
    var rplRatio float64
    var queueBalance float64
    for received := 0; received < 5; {
        select {
            case networkEthCapacity = <-networkEthCapacityChannel:
                received++
            case networkEthAssigned = <-networkEthAssignedChannel:
                received++
            case networkUtilisation = <-networkUtilisationChannel:
                received++
            case rplRatio = <-rplRatioChannel:
                received++
            case queueBalance = <-queueBalanceChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return &StakingDurationMetrics{
        DurationId: durationId,
        NetworkEthCapacity: networkEthCapacity,
        NetworkEthAssigned: networkEthAssigned,
        NetworkUtilisation: networkUtilisation,
        RplRatio: rplRatio,
        QueueBalance: queueBalance,
    }, nil

}

