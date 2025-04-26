package strings

import "testing"

func TestSnakeNameToBigCamelName(t *testing.T) {
	t.Run("user_center", func(t *testing.T) {
		name := SnakeNameToBigCamelName("user_center")
		if name != "UserCenter" {
			t.Fatalf("Name is not expectation:%s", name)
		}
	})

	t.Run("vip_channel_type", func(t *testing.T) {
		name := SnakeNameToBigCamelName("vip_channel_type")
		if name != "VipChannelType" {
			t.Fatalf("Name is not expectation:%s", name)
		}
	})

	t.Run("user_id", func(t *testing.T) {
		name := SnakeNameToBigCamelName("user_id")
		if name != "UserId" {
			t.Fatalf("Name is not expectation:%s", name)
		}
	})
}
