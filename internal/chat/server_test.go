package chat

import (
	"context"
	"testing"

	"github.com/acai-travel/tech-challenge/internal/chat/model"
	. "github.com/acai-travel/tech-challenge/internal/chat/testing"
	"github.com/acai-travel/tech-challenge/internal/pb"
	"github.com/google/go-cmp/cmp"
	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/testing/protocmp"
)

type fakeAssistant struct {
	title    string
	reply    string
	titleErr error
	replyErr error
}

func TestServer_DescribeConversation(t *testing.T) {
	ctx := context.Background()
	srv := NewServer(model.New(ConnectMongo()), nil)

	t.Run("describe existing conversation", WithFixture(func(t *testing.T, f *Fixture) {
		c := f.CreateConversation()

		out, err := srv.DescribeConversation(ctx, &pb.DescribeConversationRequest{ConversationId: c.ID.Hex()})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, want := out.GetConversation(), c.Proto()
		if !cmp.Equal(got, want, protocmp.Transform()) {
			t.Errorf("DescribeConversation() mismatch (-got +want):\n%s", cmp.Diff(got, want, protocmp.Transform()))
		}
	}))

	t.Run("describe non existing conversation should return 404", WithFixture(func(t *testing.T, f *Fixture) {
		_, err := srv.DescribeConversation(ctx, &pb.DescribeConversationRequest{ConversationId: "08a59244257c872c5943e2a2"})
		if err == nil {
			t.Fatal("expected error for non-existing conversation, got nil")
		}

		if te, ok := err.(twirp.Error); !ok || te.Code() != twirp.NotFound {
			t.Fatalf("expected twirp.NotFound error, got %v", err)
		}
	}))
}

func (f *fakeAssistant) Title(ctx context.Context, _ *model.Conversation) (string, error) {
	return f.title, f.titleErr
}
func (f *fakeAssistant) Reply(ctx context.Context, _ *model.Conversation) (string, error) {
	return f.reply, f.replyErr
}

func TestServer_StartConversation_Success(t *testing.T) {
	ctx := context.Background()

	srv := NewServer(model.New(ConnectMongo()), &fakeAssistant{
		title: "Weather in Barcelona",
		reply: "25°C and sunny",
	})

	req := &pb.StartConversationRequest{Message: "What's the weather in Barcelona?"}
	resp, err := srv.StartConversation(ctx, req)
	if err != nil {
		t.Fatalf("StartConversation error: %v", err)
	}

	if resp.GetConversationId() == "" {
		t.Fatal("expected non-empty conversation id")
	}
	if got, want := resp.GetTitle(), "Weather in Barcelona"; got != want {
		t.Fatalf("title: got %q, want %q", got, want)
	}
	if got, want := resp.GetReply(), "25°C and sunny"; got != want {
		t.Fatalf("reply: got %q, want %q", got, want)
	}

	out, err := srv.DescribeConversation(ctx, &pb.DescribeConversationRequest{
		ConversationId: resp.GetConversationId(),
	})
	if err != nil {
		t.Fatalf("DescribeConversation error: %v", err)
	}
	if got, want := out.GetConversation().GetTitle(), "Weather in Barcelona"; got != want {
		t.Fatalf("saved title: got %q, want %q", got, want)
	}
	msgs := out.GetConversation().GetMessages()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages (user + assistant), got %d", len(msgs))
	}
	if msgs[1].GetRole() != pb.Conversation_ASSISTANT || msgs[1].GetContent() != "25°C and sunny" {
		t.Fatalf("assistant message mismatch: role=%v content=%q", msgs[1].GetRole(), msgs[1].GetContent())
	}
}
