import urllib.request, json
req = urllib.request.Request('http://localhost:8080/api/v1/auth/login',
    data=json.dumps({"email":"trip_tester@test.ru","password":"test12345"}).encode('utf-8'),
    headers={'Content-Type': 'application/json'})
try:
    with urllib.request.urlopen(req) as response:
        print(json.loads(response.read().decode())['data']['tokens']['access_token'])
except Exception as e:
    print("Error:", e)
