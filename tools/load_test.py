import urllib.request
import urllib.error
import time
import threading
import random
import argparse
from concurrent.futures import ThreadPoolExecutor

# Metrics
stats = {
    "total_requests": 0,
    "success": 0,
    "errors": 0,
    "durations": []
}
lock = threading.Lock()
error_messages = set()

def send_request(target, host_header, endpoints):
    url = target + random.choice(endpoints)
    headers = {
        "Host": host_header,
        "User-Agent": "SyntheticLoadTester/1.0"
    }
    req = urllib.request.Request(url, headers=headers)
    start_time = time.time()
    try:
        with urllib.request.urlopen(req, timeout=5) as response:
            status = response.getcode()
            response.read() # Read response body
            duration = (time.time() - start_time) * 1000 # in ms
            with lock:
                stats["total_requests"] += 1
                if status == 200:
                    stats["success"] += 1
                else:
                    stats["errors"] += 1
                stats["durations"].append(duration)
    except urllib.error.HTTPError as e:
        duration = (time.time() - start_time) * 1000
        err_msg = f"HTTP {e.code}: {url}"
        with lock:
            stats["total_requests"] += 1
            stats["errors"] += 1
            stats["durations"].append(duration)
            if len(error_messages) < 10 and err_msg not in error_messages:
                error_messages.add(err_msg)
                print(f"Error: {err_msg}")
    except Exception as e:
        duration = (time.time() - start_time) * 1000
        err_msg = f"Exception {e.__class__.__name__}: {e} for {url}"
        with lock:
            stats["total_requests"] += 1
            stats["errors"] += 1
            stats["durations"].append(duration)
            if len(error_messages) < 10 and err_msg not in error_messages:
                error_messages.add(err_msg)
                print(f"Error: {err_msg}")

def worker_loop(stop_event, target, host_header, endpoints):
    while not stop_event.is_set():
        send_request(target, host_header, endpoints)
        time.sleep(0.01) # 10ms delay between requests per thread

def main():
    parser = argparse.ArgumentParser(description="Synthetic Load Test tool for Avtoplaneta API Gateway")
    parser.add_argument("--target", default="http://localhost", help="Target base URL (default: http://localhost)")
    parser.add_argument("--host-header", default="192.168.1.63", help="Host header for routing (default: 192.168.1.63)")
    parser.add_argument("--threads", type=int, default=24, help="Number of concurrent threads (default: 24)")
    parser.add_argument("--duration", type=int, default=20, help="Test duration in seconds (default: 20)")
    
    args = parser.parse_args()
    
    endpoints = ["/health", "/api/inventory", "/api/statistics"]
    
    print("Starting synthetic load test...")
    print(f"Target: {args.target}")
    print(f"Host Header: {args.host_header}")
    print(f"Threads: {args.threads}")
    print(f"Duration: {args.duration} seconds")
    print(f"Endpoints: {endpoints}")
    print("-------------------------------------------------")
    
    stop_event = threading.Event()
    start_time = time.time()
    
    with ThreadPoolExecutor(max_workers=args.threads) as executor:
        futures = [executor.submit(worker_loop, stop_event, args.target, args.host_header, endpoints) for _ in range(args.threads)]
        
        try:
            for sec in range(args.duration):
                time.sleep(1)
                with lock:
                    total = stats["total_requests"]
                    ok = stats["success"]
                    err = stats["errors"]
                    avg_dur = sum(stats["durations"]) / len(stats["durations"]) if stats["durations"] else 0
                    print(f"Elapsed: {sec+1}s | Requests: {total} | Success: {ok} | Errors: {err} | Avg Latency: {avg_dur:.2f}ms")
        except KeyboardInterrupt:
            print("\nStopping load test...")
        
        stop_event.set()
        print("Waiting for worker threads to shut down...")
    
    total_time = time.time() - start_time
    with lock:
        total = stats["total_requests"]
        ok = stats["success"]
        err = stats["errors"]
        avg_dur = sum(stats["durations"]) / len(stats["durations"]) if stats["durations"] else 0
        rps = total / total_time
        pct_success = (ok / total * 100) if total > 0 else 0
        print("\n--- Load Test Results ---")
        print(f"Total Duration: {total_time:.2f} seconds")
        print(f"Total Requests: {total}")
        print(f"Successful Requests: {ok} ({pct_success:.2f}%)")
        print(f"Failed Requests: {err}")
        print(f"Average Request Latency: {avg_dur:.2f}ms")
        print(f"Requests Per Second (RPS): {rps:.2f}")

if __name__ == "__main__":
    main()
