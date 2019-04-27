package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessagesIntegrity(t *testing.T) {
	t.Log("FatalErrorMessage")
	{
		assert.Equal(t, "TO FROM EE", FatalErrorMessage("FROM", "TO"))
	}

	t.Log("AccountCreatedMessage")
	{
		assert.Equal(t, "TO FROM AN", AccountCreatedMessage("FROM", "TO"))
	}

	t.Log("PromiseAcceptedMessage")
	{
		assert.Equal(t, "TO FROM P1", PromiseAcceptedMessage("FROM", "TO"))
	}

	t.Log("PromiseRejectedMessage")
	{
		assert.Equal(t, "TO FROM P2 REASON", PromiseRejectedMessage("FROM", "TO", "REASON"))
	}

	t.Log("CommitAcceptedMessage")
	{
		assert.Equal(t, "TO FROM C1", CommitAcceptedMessage("FROM", "TO"))
	}

	t.Log("CommitRejectedMessage")
	{
		assert.Equal(t, "TO FROM C2 REASON", CommitRejectedMessage("FROM", "TO", "REASON"))
	}

	t.Log("RollbackAcceptedMessage")
	{
		assert.Equal(t, "TO FROM R1", RollbackAcceptedMessage("FROM", "TO"))
	}

	t.Log("RollbackRejectedMessage")
	{
		assert.Equal(t, "TO FROM R2 REASON", RollbackRejectedMessage("FROM", "TO", "REASON"))
	}

	t.Log("AccountStateMessage")
	{
		assert.Equal(t, "TO FROM S0 CURRENCY t BALANCE PROMISED", AccountStateMessage("FROM", "TO", "CURRENCY", "BALANCE", "PROMISED", true))
		assert.Equal(t, "TO FROM S0 CURRENCY f BALANCE PROMISED", AccountStateMessage("FROM", "TO", "CURRENCY", "BALANCE", "PROMISED", false))
	}
}
