---
description: RESTFul API endpoint description
---

# RESTFul API

## Step to send Tx

* Run rest server

```text
./clif rest-server
```

* Execute appropriate API endpoint
  * If sending tx purpose, proceed to sign tx 
  * If query purpose, use the response
* Sign tx in local, and take the result
* Send the response above with \`send tx\` of bank API

{% api-method method="get" host="https://localhost:1317" path="/executionlayer/balance?address=friday15kvpva2u57vv6l5czehyk69s0wnq9hrkqulwfz" %}
{% api-method-summary %}
Get  balance
{% endapi-method-summary %}

{% api-method-description %}
This endpoint allows you to get HDAC balance
{% endapi-method-description %}

{% api-method-spec %}
{% api-method-request %}
{% api-method-query-parameters %}
{% api-method-parameter name="address" type="string" required=true %}
Wallet of Hdac wallet address starting from 'friday'
{% endapi-method-parameter %}
{% endapi-method-query-parameters %}
{% endapi-method-request %}

{% api-method-response %}
{% api-method-response-example httpCode=200 %}
{% api-method-response-example-description %}
Balance successfully retrieved.
{% endapi-method-response-example-description %}

```
{
    "value": 5000000
}
```
{% endapi-method-response-example %}
{% endapi-method-response %}
{% endapi-method-spec %}
{% endapi-method %}

{% api-method method="post" host="https://localhost:1317" path="/executionlayer/bond" %}
{% api-method-summary %}
Bond balance
{% endapi-method-summary %}

{% api-method-description %}
Bond balance of validator \(Experimental endpoint for WASM execution\)
{% endapi-method-description %}

{% api-method-spec %}
{% api-method-request %}
{% api-method-form-data-parameters %}
{% api-method-parameter name="amount" type="integer" required=true %}
The integer amount you want to bond
{% endapi-method-parameter %}

{% api-method-parameter name="gas\_fee" type="integer" required=true %}
Tx gas fee. If less than the required, tx will be failed
{% endapi-method-parameter %}

{% api-method-parameter name="address" type="string" required=true %}
Wallet of Hdac wallet address starting from 'friday'
{% endapi-method-parameter %}
{% endapi-method-form-data-parameters %}
{% endapi-method-request %}

{% api-method-response %}
{% api-method-response-example httpCode=200 %}
{% api-method-response-example-description %}

{% endapi-method-response-example-description %}

```
TBD
```
{% endapi-method-response-example %}
{% endapi-method-response %}
{% endapi-method-spec %}
{% endapi-method %}

{% api-method method="post" host="https://localhost:1317" path="/executionlayer/unbond" %}
{% api-method-summary %}
Unbond balance
{% endapi-method-summary %}

{% api-method-description %}
Unbond balance of validator \(Experimental endpoint for WASM execution\)
{% endapi-method-description %}

{% api-method-spec %}
{% api-method-request %}
{% api-method-form-data-parameters %}
{% api-method-parameter name="amojunt" type="integer" required=true %}
The integer amount you want to unbond
{% endapi-method-parameter %}

{% api-method-parameter name="gas\_fee" type="integer" required=true %}
Tx gas fee. If less than the required, tx will be failed
{% endapi-method-parameter %}

{% api-method-parameter name="address" type="string" required=true %}
Wallet of Hdac wallet address starting from 'friday'
{% endapi-method-parameter %}
{% endapi-method-form-data-parameters %}
{% endapi-method-request %}

{% api-method-response %}
{% api-method-response-example httpCode=200 %}
{% api-method-response-example-description %}

{% endapi-method-response-example-description %}

```

```
{% endapi-method-response-example %}
{% endapi-method-response %}
{% endapi-method-spec %}
{% endapi-method %}

{% api-method method="post" host="https://localhost:1317" path="/executionlayer/transfer" %}
{% api-method-summary %}
Transfer
{% endapi-method-summary %}

{% api-method-description %}
Transfer token from one to another \(under construction\)
{% endapi-method-description %}

{% api-method-spec %}
{% api-method-request %}
{% api-method-form-data-parameters %}
{% api-method-parameter name="status" type="integer" required=true %}
OK - 0 / Fail - 1
{% endapi-method-parameter %}

{% api-method-parameter name="recipient\_address" type="string" required=true %}
Receiver address
{% endapi-method-parameter %}

{% api-method-parameter name="sender\_address" type="string" required=true %}
Address of token sender
{% endapi-method-parameter %}

{% api-method-parameter name="amount" type="integer" required=true %}
Amount of token you want to send
{% endapi-method-parameter %}
{% endapi-method-form-data-parameters %}
{% endapi-method-request %}

{% api-method-response %}
{% api-method-response-example httpCode=200 %}
{% api-method-response-example-description %}

{% endapi-method-response-example-description %}

```
TBD
```
{% endapi-method-response-example %}
{% endapi-method-response %}
{% endapi-method-spec %}
{% endapi-method %}

