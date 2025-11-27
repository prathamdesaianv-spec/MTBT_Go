# NSE MTBT (Multicast Tick-by-Tick) Receiver

A high-performance Go application that receives real-time market data from NSE's FO (Futures & Options) segment using UDP multicast.

## Overview

This application connects to **36 simultaneous UDP multicast feeds** (18 FO streams × 2 sources each) and receives tick-by-tick order and trade data from the National Stock Exchange of India.

### Features

- ✅ Connects to all 18 FO market data streams
- ✅ Dual source connectivity (Source 1 + Source 2 delayed feed)
- ✅ 36 concurrent goroutines for parallel data reception
- ✅ Protocol-compliant parsing based on MTBT API v6.7
- ✅ Real-time statistics monitoring
- ✅ Support for multiple message types:
  - Order Messages (New/Modify/Cancel)
  - Trade Messages
  - Spread Order Messages
  - Spread Trade Messages
  - Trade Cancel Messages
  - Heartbeat Messages

## Architecture

### FO Streams Configuration

| Stream | Stream ID | Source 1 IP:Port | Source 2 IP:Port | Bandwidth |
|--------|-----------|------------------|------------------|-----------|
| FO_B   | 1         | 239.70.70.41:17741 | 239.70.70.31:10831 | 20 Mbps |
| FO_C   | 2         | 239.70.70.42:17742 | 239.70.70.32:10832 | 6 Mbps |
| FO_D   | 3         | 239.70.70.43:17743 | 239.70.70.33:10833 | 40 Mbps |
| FO_E   | 4         | 239.70.70.44:17744 | 239.70.70.34:10834 | 18 Mbps |
| FO_F   | 5         | 239.70.70.45:17745 | 239.70.70.35:10835 | 40 Mbps |
| FO_G   | 6         | 239.70.70.46:17746 | 239.70.70.36:10836 | 18 Mbps |
| FO_H   | 7         | 239.70.70.47:17747 | 239.70.70.37:10837 | 40 Mbps |
| FO_I   | 8         | 239.70.70.48:17748 | 239.70.70.38:10838 | 40 Mbps |
| FO_J   | 9         | 239.70.70.49:17749 | 239.70.70.39:10839 | 40 Mbps |
| FO_K   | 10        | 239.70.70.50:17750 | 239.70.70.40:10840 | 40 Mbps |
| FO_L   | 11        | 239.70.70.61:17761 | 239.70.70.71:10871 | 40 Mbps |
| FO_M   | 12        | 239.70.70.62:17762 | 239.70.70.72:10872 | 40 Mbps |
| FO_N   | 13        | 239.70.70.63:17763 | 239.70.70.73:10873 | 40 Mbps |
| FO_O   | 14        | 239.70.70.64:17764 | 239.70.70.74:10874 | 40 Mbps |
| FO_P   | 15        | 239.70.70.65:17765 | 239.70.70.75:10875 | 40 Mbps |
| FO_Q   | 16        | 239.70.70.66:17766 | 239.70.70.76:10876 | 40 Mbps |
| FO_R   | 17        | 239.70.70.67:17767 | 239.70.70.77:10877 | 40 Mbps |
| FO_S   | 18        | 239.70.70.68:17768 | 239.70.70.78:10878 | 40 Mbps |

**Total: 36 concurrent connections**

## Data Structures

Based on NSE MTBT API Specification v6.7:

### Stream Header (8 bytes)
- Message Length (2 bytes) - Total packet size
- Stream ID (2 bytes) - Identifies the stream
- Sequence Number (4 bytes) - Packet sequence (UINT, max 4,294,967,295)

### Message Types
- `N` - New Order
- `M` - Order Modification
- `X` - Order Cancellation
- `T` - Trade
- `G` - New Spread Order
- `H` - Spread Order Modification
- `J` - Spread Order Cancellation
- `K` - Spread Trade
- `C` - Trade Cancel
- `Z` - Heartbeat

## Requirements

- **Go 1.23.3+** (installed)
- **Network Configuration**:
  - Multicast routing enabled
  - Access to NSE multicast groups (239.70.70.x)
  - Proper network permissions
- **NSE Membership**: Valid NSE membership with market data access
- **Master Files**: Contract information files from NSE extranet

## Installation

The Go language is already installed. To build the MTBT receiver:

```bash
cd /teamspace/studios/this_studio/MTBT_Go
go build -o mtbt_receiver mtbt_receiver.go
```

## Usage

### Run the receiver:

```bash
./mtbt_receiver
```

### Expected Output:

```
========================================
NSE MTBT (Multicast Tick-by-Tick) Receiver
FO Segment - All 36 Connections
========================================
Launched 36 goroutines for 18 FO streams (Source1 + Source2)
Note: You need proper network access and permissions to receive NSE market data
MTBT Receiver is running. Press Ctrl+C to stop.

[FO_B-Source1] Starting listener on 239.70.70.41:17741
[FO_B-Source2] Starting listener on 239.70.70.31:10831
[FO_C-Source1] Starting listener on 239.70.70.42:17742
...
```

### Statistics Output (every 30 seconds):

```
========== MTBT Statistics Summary ==========
[FO_B-Source1] Packets: 1523 | Seq: 1523 | Orders: 1200 | Trades: 323 | Errors: 0 | Last: 1s
[FO_C-Source1] Packets: 892 | Seq: 892 | Orders: 700 | Trades: 192 | Errors: 0 | Last: 2s
...
=============================================
```

## Network Configuration

### Enable Multicast (if needed):

```bash
# Check if multicast route exists
route -n | grep 224.0.0.0

# Add multicast route (if missing)
sudo route add -net 224.0.0.0 netmask 240.0.0.0 dev eth0
```

### Increase UDP Buffer Sizes (recommended):

```bash
# Temporary
sudo sysctl -w net.core.rmem_max=134217728
sudo sysctl -w net.core.rmem_default=134217728

# Permanent (add to /etc/sysctl.conf)
net.core.rmem_max=134217728
net.core.rmem_default=134217728
```

## Performance Tuning

The application implements several optimizations from NSE guidelines:

1. **8MB UDP read buffers** per connection
2. **Concurrent goroutines** for parallel processing
3. **Little-endian byte order** parsing
4. **Reduced logging** for high-frequency streams (every 100 packets)

## Important Notes

⚠️ **Production Considerations:**

1. **Access Required**: You need authorized access to NSE market data feeds
2. **Network Setup**: Proper multicast routing and firewall configuration
3. **Master Files**: Download contract information files from NSE extranet
4. **Data Rights**: Comply with NSE data usage policies and regulations
5. **Testing**: Test on mock/simulation environment before production
6. **Error Handling**: Implement reconnection logic for production use
7. **Persistence**: Add database storage for order book reconstruction

## References

- **MTBT API Specification v6.7** - `/NSE docs/MTBT_FO_NNF_PROTOCL_6.7_0 (1).pdf`
- **NSE Market Data Circular** - `/NSE docs/MSD67677.pdf`
- NSE Website: https://www.nseindia.com/

## License

See LICENSE file in repository.

## Contact

For NSE technical support:
- Toll Free: 1800-266-0050 (Option 1)
- Email: msm@nse.co.in
