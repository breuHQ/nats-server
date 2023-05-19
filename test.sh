for idx in 0{1..20}
do
    curl --location 'http://localhost:1323/sdk/message' \
    --header 'Content-Type: application/json' \
    --header 'Authorization: Bearer eeb1bc1d-b5a2-43cb-8521-7355d5832dc0' \
    --data '{
        "service_id": "853b7275-c640-462d-961a-625105faf5e5",
        "operation_id": "CreateMessage",
        "req_body": {
            "user_id": "7f47fa0c-3035-4fbf-aff1-c3f7b639c7df",
            "Body": "hello from postman",
            "To": "+923214930471",
            "From": "+19034598701"
        },
        "path_params": {
            "AccountSid": "AC9f560ea30baaaf8013e4e44284eb6769"
        },
        "query_params": {
        }
    }'
done