package executionlayer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceFromBech32ToHex(t *testing.T) {
	testInput := `[{"name":"method","value":{"cl_type":{"simple_type":"STRING"},"value":{"str_value":"set_swap_hash"}}},` +
		`{"name":"hash","value":{"cl_type":{"simple_type":"KEY"},"value":{"key":{"hash":{"hash":"fridaycontracthash1fh7vqy3zp945f0xel3h7x7sj6rq5hw509q2jmdsfndqh8ygcj4mqcgh9n7"}}}}},` +
		`{"name":"uref","value":{"cl_type":{"simple_type":"KEY"},"value":{"key":{"uref":{"uref":"fridaycontracturef1zqaevn0n0haygwq9dmqk0came8lmxqa6fp876pvsj54mm0pnycjssj50u9"}}}}},` +
		`{"name":"address","value":{"cl_type":{"list_type":{"inner":{"simple_type":"U8"}}},"value":{"bytes_value":"friday1k568qc388n6x5ks8hkwly2q9ruepns8rr9sgqyjxk9cy6a2qq8gs4v2kpm"}}},` +
		`{"name":"address_as_string","value":{"cl_type":{"simple_type":"STRING"},"value":{"str_value":"friday1k568qc388n6x5ks8hkwly2q9ruepns8rr9sgqyjxk9cy6a2qq8gs4v2kpm"}}}]`

	res, _ := ReplaceFromBech32ToHex(testInput)
	fmt.Println(res)

	expectedRes := `[{"name":"method","value":{"cl_type":{"simple_type":"STRING"},"value":{"str_value":"set_swap_hash"}}},` +
		`{"name":"hash","value":{"cl_type":{"simple_type":"KEY"},"value":{"key":{"hash":{"hash":"TfzAEiIJa0S82fxv43oS0MFLuo8oFS22CZtBc5EYlXY="}}}}},` +
		`{"name":"uref","value":{"cl_type":{"simple_type":"KEY"},"value":{"key":{"uref":{"uref":"EDuWTfN9+kQ4BW7BZ+O7yf+zA7pIT+0FkJUrvbwzJiU="}}}}},` +
		`{"name":"address","value":{"cl_type":{"list_type":{"inner":{"simple_type":"U8"}}},"value":{"bytes_value":"tTRwYic89GpaB72d8igFHzIZwOMZYIASRrFwTXVAAdE="}}},` +
		`{"name":"address_as_string","value":{"cl_type":{"simple_type":"STRING"},"value":{"str_value":"friday1k568qc388n6x5ks8hkwly2q9ruepns8rr9sgqyjxk9cy6a2qq8gs4v2kpm"}}}]`

	assert.Equal(t, res, expectedRes)
}
