#!/usr/bin/env python3
"""Test helper for MsgPack/JSON content type support."""
import subprocess
import json
import sys

# Try importing msgpack
try:
    import msgpack
    HAS_MSGPACK = True
except ImportError:
    HAS_MSGPACK = False
    print("msgpack not installed. Install with: pip install msgpack")
    sys.exit(1)

# The server expects:
# - Content-Type: application/json or application/msgpack
# - Body: JSON-encoded object with app_id, user_id, action_name, action_param

payload = {
    "app_id": 1,
    "user_id": 1,
    "action_name": "test",
    "action_param": '{"key": "value"}'
}

# Test 1: JSON
print("Testing Content-Type: application/json")
json_bytes = json.dumps(payload).encode('utf-8')
result = subprocess.run([
    'curl', '-s', '-X', 'POST', 'http://localhost:7781/actions/run',
    '-H', 'Content-Type: application/json',
    '-d', json.dumps(payload)
], capture_output=True, text=True)
print(f"Status: {result.returncode}")
print(f"Response: {result.stdout[:200]}")
print()

# Test 2: MsgPack (binary)
if HAS_MSGPACK:
    print("Testing Content-Type: application/msgpack")
    msgpack_bytes = msgpack.packb(payload, default=str)
    result = subprocess.run([
        'curl', '-s', '-X', 'POST', 'http://localhost:7781/actions/run',
        '-H', 'Content-Type: application/msgpack',
        '--data-binary', f'@-;type=application/msgpack',
    ], input=msgpack_bytes, capture_output=True)
    print(f"Status: {result.returncode}")
    print(f"Response: {result.stdout[:200]}")
    print()

# Test 3: Unsupported content type
print("Testing Content-Type: application/xml (should return 415)")
result = subprocess.run([
    'curl', '-s', '-X', 'POST', 'http://localhost:7781/actions/run',
    '-H', 'Content-Type: application/xml',
    '-d', '<xml></xml>'
], capture_output=True, text=True)
print(f"Status: {result.returncode}")
print(f"Response: {result.stdout[:200]}")
