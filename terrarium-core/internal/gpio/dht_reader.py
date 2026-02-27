import sys
import json
import board
import adafruit_dht

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No pin provided"}))
        sys.exit(1)
        
    pin_num = int(sys.argv[1])
    # Dynamically select board.Dx where x is the BCM pin
    pin_board = getattr(board, f"D{pin_num}", None)
    if pin_board is None:
        print(json.dumps({"error": f"Invalid board pin D{pin_num}"}))
        sys.exit(1)
        
    try:
        dhtDevice = adafruit_dht.DHT22(pin_board, use_pulseio=False)
        temperature_c = dhtDevice.temperature
        humidity = dhtDevice.humidity
        dhtDevice.exit()
        
        if temperature_c is None or humidity is None:
            print(json.dumps({"error": "Failed to read sensor: returned None"}))
        else:
            print(json.dumps({"temperature": temperature_c, "humidity": humidity}))
    except RuntimeError as error:
        # Errors happen fairly often, DHT's are hard to read
        print(json.dumps({"error": error.args[0]}))
        sys.exit(1)
    except Exception as error:
        dhtDevice.exit()
        print(json.dumps({"error": str(error)}))
        sys.exit(1)

if __name__ == "__main__":
    main()
