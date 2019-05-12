package actor

import (
	"testing"

	system "github.com/jancajthaml-openbank/actor-system"

	"github.com/stretchr/testify/assert"
)

func TestMessagesIntegrity(t *testing.T) {
	context := system.Context{
		Sender: system.Coordinates{
			Name:   "FROM_NAME",
			Region: "FROM_REGION",
		},
		Receiver: system.Coordinates{
			Name:   "TO_NAME",
			Region: "TO_REGION",
		},
	}

	t.Log("FatalErrorMessage")
	{
		assert.Equal(t, "FROM_REGION TO_REGION FROM_NAME TO_NAME EE", FatalErrorMessage(context))
	}

	t.Log("AccountCreatedMessage")
	{
		assert.Equal(t, "FROM_REGION TO_REGION FROM_NAME TO_NAME AN", AccountCreatedMessage(context))
	}

	t.Log("PromiseAcceptedMessage")
	{
		assert.Equal(t, "FROM_REGION TO_REGION FROM_NAME TO_NAME P1", PromiseAcceptedMessage(context))
	}

	t.Log("PromiseRejectedMessage")
	{
		assert.Equal(t, "FROM_REGION TO_REGION FROM_NAME TO_NAME P2 REASON", PromiseRejectedMessage(context, "REASON"))
	}

	t.Log("CommitAcceptedMessage")
	{
		assert.Equal(t, "FROM_REGION TO_REGION FROM_NAME TO_NAME C1", CommitAcceptedMessage(context))
	}

	t.Log("CommitRejectedMessage")
	{
		assert.Equal(t, "FROM_REGION TO_REGION FROM_NAME TO_NAME C2 REASON", CommitRejectedMessage(context, "REASON"))
	}

	t.Log("RollbackAcceptedMessage")
	{
		assert.Equal(t, "FROM_REGION TO_REGION FROM_NAME TO_NAME R1", RollbackAcceptedMessage(context))
	}

	t.Log("RollbackRejectedMessage")
	{
		assert.Equal(t, "FROM_REGION TO_REGION FROM_NAME TO_NAME R2 REASON", RollbackRejectedMessage(context, "REASON"))
	}

	t.Log("AccountStateMessage")
	{
		assert.Equal(t, "FROM_REGION TO_REGION FROM_NAME TO_NAME S0 CURRENCY t BALANCE PROMISED", AccountStateMessage(context, "CURRENCY", "BALANCE", "PROMISED", true))
		assert.Equal(t, "FROM_REGION TO_REGION FROM_NAME TO_NAME S0 CURRENCY f BALANCE PROMISED", AccountStateMessage(context, "CURRENCY", "BALANCE", "PROMISED", false))
	}
}
