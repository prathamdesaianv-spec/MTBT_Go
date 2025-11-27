# Quick Start Guide - NSE MTBT Receiver

## What Was Built

A complete Go application that connects to **36 simultaneous NSE market data feeds** for the FO (Futures & Options) segment:
- 18 FO streams (FO_B through FO_S)
- Each stream has 2 sources (Source1 + Source2 delayed feed)
- **Total: 36 concurrent UDP multicast connections**

## Files Created

```
MTBT_Go/
├── mtbt_receiver.go      # Main Go source code (36 goroutines)
├── mtbt_receiver         # Compiled executable (3.1MB)
├── MTBT_README.md        # Detailed documentation
├── hello_world.go        # Simple Go test program
└── text_convert/         # PDF to text conversion utilities
    └── output/
        ├── MSD67677.txt                        # Market data circular
        └── MTBT_FO_NNF_PROTOCL_6.7_0 (1).txt # API specification
```

## How to Run

### 1. Basic Execution
```bash
cd /teamspace/studios/this_studio/MTBT_Go
export PATH=$PATH:/usr/local/go/bin
./mtbt_receiver
```

### 2. Run with Output Logging
```bash
./mtbt_receiver 2>&1 | tee mtbt_output.log
```

### 3. Run in Background
```bash
nohup ./mtbt_receiver > mtbt.log 2>&1 &
```

### 4. Stop the Program
Press `Ctrl+C` or if running in background:
```bash
pkill mtbt_receiver
```

## What It Does

### Connections (All 36)
The program launches 36 goroutines, each listening on a different UDP multicast address:

**Example connections:**
- FO_B-Source1: 239.70.70.41:17741
- FO_B-Source2: 239.70.70.31:10831
- FO_C-Source1: 239.70.70.42:17742
- FO_C-Source2: 239.70.70.32:10832
- ... (and 32 more)

### Data Received
Each goroutine receives and processes:
- **Order Messages**: New orders, modifications, cancellations
- **Trade Messages**: Order executions
- **Spread Orders/Trades**: Spread contract trading
- **Trade Cancellations**: Cancelled trades
- **Heartbeat Messages**: Keep-alive signals

### Real-time Statistics
Every 30 seconds, you'll see:
```
========== MTBT Statistics Summary ==========
[FO_B-Source1] Packets: 1523 | Seq: 1523 | Orders: 1200 | Trades: 323 | Errors: 0
[FO_C-Source1] Packets: 892 | Seq: 892 | Orders: 700 | Trades: 192 | Errors: 0
...
```

## Protocol Details

### Packet Structure (Little Endian, Pragma Pack 1)

**Header (8 bytes):**
```
[0-1] Message Length (int16)
[2-3] Stream ID (int16)
[4-7] Sequence Number (uint32)
```

**Message Types:**
- `N` = New Order
- `M` = Modify Order
- `X` = Cancel Order
- `T` = Trade
- `G` = New Spread Order
- `H` = Modify Spread Order
- `J` = Cancel Spread Order
- `K` = Spread Trade
- `C` = Trade Cancel
- `Z` = Heartbeat

## Network Requirements

⚠️ **Important**: This program requires:

1. **Network Access**: Access to NSE's multicast network
2. **Multicast Routing**: Properly configured multicast routes
3. **Firewall Rules**: Allow UDP traffic on ports 10831-10878 and 17741-17768
4. **NSE Authorization**: Valid NSE membership and data subscription

### Test Network Connectivity
```bash
# Check if you can reach multicast addresses
ping 239.70.70.41

# Check multicast routing
route -n | grep 224.0.0.0

# Monitor network traffic
sudo tcpdump -i any 'multicast and (port 17741 or port 17742)'
```

## Performance Tuning

The application is pre-configured with:
- 8MB UDP receive buffers per connection
- 64KB packet buffers
- Reduced logging for high-frequency streams
- Little-endian byte parsing

### Optional System Tuning:
```bash
# Increase UDP buffer sizes
sudo sysctl -w net.core.rmem_max=134217728
sudo sysctl -w net.core.rmem_default=134217728
```

## Expected Bandwidth

Total approximate bandwidth for all 36 connections:
- **Source 1 (18 streams)**: ~500 Mbps
- **Source 2 (18 delayed feeds)**: ~500 Mbps
- **Total**: ~1 Gbps during peak trading hours

Individual stream bandwidth varies from 6 Mbps to 40 Mbps.

## Troubleshooting

### No Data Received
```
✗ Check network connectivity
✗ Verify NSE market hours (9:15 AM - 3:30 PM IST)
✗ Confirm multicast routing is enabled
✗ Check firewall rules
```

### High Error Count
```
✗ Network congestion - increase buffer sizes
✗ Packet loss - check network quality
✗ System overload - reduce logging frequency
```

### Build Errors
```bash
# Rebuild from source
cd /teamspace/studios/this_studio/MTBT_Go
go clean
go build -o mtbt_receiver mtbt_receiver.go
```

## Development

### Modify the Code
```bash
# Edit the source
nano mtbt_receiver.go

# Rebuild
go build -o mtbt_receiver mtbt_receiver.go

# Run
./mtbt_receiver
```

### Add Features
Common enhancements:
- Database persistence (PostgreSQL/TimescaleDB)
- Order book reconstruction
- WebSocket API for clients
- Metrics export (Prometheus)
- Alert system for anomalies

## References

- **API Spec**: `NSE docs/MTBT_FO_NNF_PROTOCL_6.7_0 (1).pdf`
- **Market Data Circular**: `NSE docs/MSD67677.pdf`
- **Detailed README**: `MTBT_README.md`
- **NSE Website**: https://www.nseindia.com/

## Support

NSE Technical Support:
- **Toll Free**: 1800-266-0050 (Option 1)
- **Email**: msm@nse.co.in

---

**Note**: This application is for educational/development purposes. Ensure compliance with NSE regulations and data usage policies before production deployment.
