package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// MTBT Protocol Structures based on NSE API Specification v6.7

// StreamHeader represents the packet header for MTBT feed
// Pragma Pack 1, Little Endian
type StreamHeader struct {
	MsgLen     int16  // Message Length
	StreamID   int16  // Stream ID
	SequenceNo uint32 // Sequence Number (UINT for FO segment)
}

// OrderMessage represents Order data (New/Modify/Cancel)
type OrderMessage struct {
	MessageType byte    // 'N'=New, 'M'=Modify, 'X'=Cancel
	Timestamp   int64   // Time in nanoseconds from 01-Jan-1980
	OrderID     float64 // Day Unique Order Reference Number
	Token       int32   // Unique Contract Identifier
	OrderType   byte    // 'B'=Buy, 'S'=Sell
	Price       int32   // Price in paise (divide by 100 for Rupees)
	Quantity    int32   // Order Quantity
}

// TradeMessage represents Trade execution data
type TradeMessage struct {
	MessageType  byte    // 'T'=Trade
	Timestamp    int64   // Time in nanoseconds from 01-Jan-1980
	BuyOrderID   float64 // Buy-Side Order ID
	SellOrderID  float64 // Sell-Side Order ID
	Token        int32   // Unique Contract Identifier
	TradePrice   int32   // Trade Price in paise (divide by 100 for Rupees)
	TradeQty     int32   // Trade Quantity
}

// SpreadOrderMessage represents Spread Order data
type SpreadOrderMessage struct {
	MessageType byte    // 'G'=New, 'H'=Modify, 'J'=Cancel
	Timestamp   int64   // Time in nanoseconds from 01-Jan-1980
	OrderID     float64 // Day Unique Order Reference Number
	Token       int32   // Unique Contract Identifier
	OrderType   byte    // 'B'=Buy, 'S'=Sell
	Price       int32   // Price difference in paise
	Quantity    int32   // Spread Order Quantity
}

// SpreadTradeMessage represents Spread Trade execution data
type SpreadTradeMessage struct {
	MessageType byte    // 'K'=Spread Trade
	Timestamp   int64   // Time in nanoseconds from 01-Jan-1980
	BuyOrderID  float64 // Buy-Side Order ID
	SellOrderID float64 // Sell-Side Order ID
	Token       int32   // Unique Contract Identifier
	TradePrice  int32   // Trade Price in paise
	Quantity    int32   // Spread Trade Quantity
}

// TradeCancelMessage represents Trade Cancellation
type TradeCancelMessage struct {
	MessageType  byte    // 'C'=Trade Cancel
	Timestamp    int64   // Time in nanoseconds from 01-Jan-1980
	BuyOrderID   float64 // Buy-Side Order ID
	SellOrderID  float64 // Sell-Side Order ID
	Token        int32   // Unique Contract Identifier
	TradePrice   int32   // Trade Price in paise
	TradeQty     int32   // Trade Quantity
}

// HeartbeatMessage represents Heartbeat message
type HeartbeatMessage struct {
	MessageType byte   // 'Z'=Heartbeat
	LastSeqNo   uint32 // Last sent data sequence number
}

// FOStream represents a FO market data stream configuration
type FOStream struct {
	StreamName    string
	StreamID      int
	Source1IP     string
	Source1Port   int
	Source2IP     string
	Source2Port   int
	ApproxBandwidth string
}

// Statistics for monitoring
type StreamStats struct {
	mu              sync.Mutex
	PacketsReceived uint64
	LastSequence    uint32
	OrderCount      uint64
	TradeCount      uint64
	ErrorCount      uint64
	LastUpdate      time.Time
}

var (
	// All 18 FO Streams with Source 1 and Source 2 configurations (Total 36 connections)
	foStreams = []FOStream{
		{StreamName: "FO_B", StreamID: 1, Source1IP: "239.70.70.41", Source1Port: 17741, Source2IP: "239.70.70.31", Source2Port: 10831, ApproxBandwidth: "20 Mbps"},
		{StreamName: "FO_C", StreamID: 2, Source1IP: "239.70.70.42", Source1Port: 17742, Source2IP: "239.70.70.32", Source2Port: 10832, ApproxBandwidth: "6 Mbps"},
		{StreamName: "FO_D", StreamID: 3, Source1IP: "239.70.70.43", Source1Port: 17743, Source2IP: "239.70.70.33", Source2Port: 10833, ApproxBandwidth: "40 Mbps"},
		{StreamName: "FO_E", StreamID: 4, Source1IP: "239.70.70.44", Source1Port: 17744, Source2IP: "239.70.70.34", Source2Port: 10834, ApproxBandwidth: "18 Mbps"},
		{StreamName: "FO_F", StreamID: 5, Source1IP: "239.70.70.45", Source1Port: 17745, Source2IP: "239.70.70.35", Source2Port: 10835, ApproxBandwidth: "40 Mbps"},
		{StreamName: "FO_G", StreamID: 6, Source1IP: "239.70.70.46", Source1Port: 17746, Source2IP: "239.70.70.36", Source2Port: 10836, ApproxBandwidth: "18 Mbps"},
		{StreamName: "FO_H", StreamID: 7, Source1IP: "239.70.70.47", Source1Port: 17747, Source2IP: "239.70.70.37", Source2Port: 10837, ApproxBandwidth: "40 Mbps"},
		{StreamName: "FO_I", StreamID: 8, Source1IP: "239.70.70.48", Source1Port: 17748, Source2IP: "239.70.70.38", Source2Port: 10838, ApproxBandwidth: "40 Mbps"},
		{StreamName: "FO_J", StreamID: 9, Source1IP: "239.70.70.49", Source1Port: 17749, Source2IP: "239.70.70.39", Source2Port: 10839, ApproxBandwidth: "40 Mbps"},
		{StreamName: "FO_K", StreamID: 10, Source1IP: "239.70.70.50", Source1Port: 17750, Source2IP: "239.70.70.40", Source2Port: 10840, ApproxBandwidth: "40 Mbps"},
		{StreamName: "FO_L", StreamID: 11, Source1IP: "239.70.70.61", Source1Port: 17761, Source2IP: "239.70.70.71", Source2Port: 10871, ApproxBandwidth: "40 Mbps"},
		{StreamName: "FO_M", StreamID: 12, Source1IP: "239.70.70.62", Source1Port: 17762, Source2IP: "239.70.70.72", Source2Port: 10872, ApproxBandwidth: "40 Mbps"},
		{StreamName: "FO_N", StreamID: 13, Source1IP: "239.70.70.63", Source1Port: 17763, Source2IP: "239.70.70.73", Source2Port: 10873, ApproxBandwidth: "40 Mbps"},
		{StreamName: "FO_O", StreamID: 14, Source1IP: "239.70.70.64", Source1Port: 17764, Source2IP: "239.70.70.74", Source2Port: 10874, ApproxBandwidth: "40 Mbps"},
		{StreamName: "FO_P", StreamID: 15, Source1IP: "239.70.70.65", Source1Port: 17765, Source2IP: "239.70.70.75", Source2Port: 10875, ApproxBandwidth: "40 Mbps"},
		{StreamName: "FO_Q", StreamID: 16, Source1IP: "239.70.70.66", Source1Port: 17766, Source2IP: "239.70.70.76", Source2Port: 10876, ApproxBandwidth: "40 Mbps"},
		{StreamName: "FO_R", StreamID: 17, Source1IP: "239.70.70.67", Source1Port: 17767, Source2IP: "239.70.70.77", Source2Port: 10877, ApproxBandwidth: "40 Mbps"},
		{StreamName: "FO_S", StreamID: 18, Source1IP: "239.70.70.68", Source1Port: 17768, Source2IP: "239.70.70.78", Source2Port: 10878, ApproxBandwidth: "40 Mbps"},
	}

	streamStats = make(map[string]*StreamStats)
	statsMutex  sync.Mutex
)

// parseStreamHeader parses the binary packet header
func parseStreamHeader(data []byte) (*StreamHeader, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("insufficient data for header: %d bytes", len(data))
	}

	header := &StreamHeader{
		MsgLen:     int16(binary.LittleEndian.Uint16(data[0:2])),
		StreamID:   int16(binary.LittleEndian.Uint16(data[2:4])),
		SequenceNo: binary.LittleEndian.Uint32(data[4:8]),
	}

	return header, nil
}

// processMessage processes different message types
func processMessage(data []byte, streamName string, source string) error {
	if len(data) < 1 {
		return fmt.Errorf("empty message data")
	}

	msgType := data[0]
	stats := getStreamStats(streamName, source)

	switch msgType {
	case 'N', 'M', 'X': // Order Messages
		stats.mu.Lock()
		stats.OrderCount++
		stats.mu.Unlock()
		log.Printf("[%s-%s] Order Message: Type=%c, DataLen=%d", streamName, source, msgType, len(data))

	case 'T': // Trade Message
		stats.mu.Lock()
		stats.TradeCount++
		stats.mu.Unlock()
		log.Printf("[%s-%s] Trade Message: Type=%c, DataLen=%d", streamName, source, msgType, len(data))

	case 'G', 'H', 'J': // Spread Order Messages
		stats.mu.Lock()
		stats.OrderCount++
		stats.mu.Unlock()
		log.Printf("[%s-%s] Spread Order Message: Type=%c, DataLen=%d", streamName, source, msgType, len(data))

	case 'K': // Spread Trade Message
		stats.mu.Lock()
		stats.TradeCount++
		stats.mu.Unlock()
		log.Printf("[%s-%s] Spread Trade Message: Type=%c, DataLen=%d", streamName, source, msgType, len(data))

	case 'C': // Trade Cancel Message
		log.Printf("[%s-%s] Trade Cancel Message: Type=%c, DataLen=%d", streamName, source, msgType, len(data))

	case 'Z': // Heartbeat Message
		log.Printf("[%s-%s] Heartbeat Message", streamName, source)

	default:
		return fmt.Errorf("unknown message type: %c (0x%02X)", msgType, msgType)
	}

	return nil
}

// getStreamStats gets or creates stream statistics
func getStreamStats(streamName, source string) *StreamStats {
	key := fmt.Sprintf("%s-%s", streamName, source)
	statsMutex.Lock()
	defer statsMutex.Unlock()

	if streamStats[key] == nil {
		streamStats[key] = &StreamStats{
			LastUpdate: time.Now(),
		}
	}
	return streamStats[key]
}

// listenMulticast listens on a UDP multicast address
func listenMulticast(stream FOStream, multicastIP string, port int, source string, wg *sync.WaitGroup) {
	defer wg.Done()

	streamID := fmt.Sprintf("%s-%s", stream.StreamName, source)
	log.Printf("[%s] Starting listener on %s:%d", streamID, multicastIP, port)

	// Parse multicast address
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", multicastIP, port))
	if err != nil {
		log.Printf("[%s] Error resolving address: %v", streamID, err)
		return
	}

	// Create UDP connection
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Printf("[%s] Error creating multicast listener: %v", streamID, err)
		return
	}
	defer conn.Close()

	// Set read buffer size (recommended in tuning guidelines)
	if err := conn.SetReadBuffer(8 * 1024 * 1024); err != nil { // 8MB buffer
		log.Printf("[%s] Warning: could not set read buffer: %v", streamID, err)
	}

	log.Printf("[%s] Successfully listening on %s:%d", streamID, multicastIP, port)

	buffer := make([]byte, 65536) // 64KB buffer for UDP packets
	stats := getStreamStats(stream.StreamName, source)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			stats.mu.Lock()
			stats.ErrorCount++
			stats.mu.Unlock()
			log.Printf("[%s] Error reading from UDP: %v", streamID, err)
			continue
		}

		// Update stats
		stats.mu.Lock()
		stats.PacketsReceived++
		stats.LastUpdate = time.Now()
		stats.mu.Unlock()

		// Parse header
		header, err := parseStreamHeader(buffer[:n])
		if err != nil {
			log.Printf("[%s] Error parsing header: %v", streamID, err)
			continue
		}

		stats.mu.Lock()
		stats.LastSequence = header.SequenceNo
		stats.mu.Unlock()

		// Process message data (after 8-byte header)
		if n > 8 {
			if err := processMessage(buffer[8:n], stream.StreamName, source); err != nil {
				log.Printf("[%s] Error processing message: %v", streamID, err)
			}
		}

		// Log packet info (reduce frequency for high-volume streams)
		if stats.PacketsReceived%100 == 0 {
			log.Printf("[%s] Packets: %d, Seq: %d, Orders: %d, Trades: %d, Errors: %d",
				streamID, stats.PacketsReceived, header.SequenceNo,
				stats.OrderCount, stats.TradeCount, stats.ErrorCount)
		}
	}
}

// printStatistics prints periodic statistics
func printStatistics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("========== MTBT Statistics Summary ==========")
		statsMutex.Lock()
		for key, stats := range streamStats {
			stats.mu.Lock()
			log.Printf("[%s] Packets: %d | Seq: %d | Orders: %d | Trades: %d | Errors: %d | Last: %s",
				key, stats.PacketsReceived, stats.LastSequence,
				stats.OrderCount, stats.TradeCount, stats.ErrorCount,
				time.Since(stats.LastUpdate).Round(time.Second))
			stats.mu.Unlock()
		}
		statsMutex.Unlock()
		log.Println("=============================================")
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("========================================")
	log.Println("NSE MTBT (Multicast Tick-by-Tick) Receiver")
	log.Println("FO Segment - All 36 Connections")
	log.Println("========================================")

	var wg sync.WaitGroup

	// Launch goroutines for all 36 connections (18 streams x 2 sources)
	totalConnections := 0
	for _, stream := range foStreams {
		// Source 1
		wg.Add(1)
		totalConnections++
		go listenMulticast(stream, stream.Source1IP, stream.Source1Port, "Source1", &wg)

		// Source 2 (Delayed Feed)
		wg.Add(1)
		totalConnections++
		go listenMulticast(stream, stream.Source2IP, stream.Source2Port, "Source2", &wg)

		// Small delay to avoid overwhelming system
		time.Sleep(10 * time.Millisecond)
	}

	log.Printf("Launched %d goroutines for %d FO streams (Source1 + Source2)", totalConnections, len(foStreams))
	log.Println("Note: You need proper network access and permissions to receive NSE market data")

	// Start statistics printer
	go printStatistics()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("MTBT Receiver is running. Press Ctrl+C to stop.")
	<-sigChan

	log.Println("\nShutdown signal received. Stopping all listeners...")
	// Note: In production, implement graceful shutdown with context cancellation
	log.Println("Goodbye!")
}
