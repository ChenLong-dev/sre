package dao

import (
	"context"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/stretchr/testify/assert"

	"rulai/models/entity"
)

func Test_Dao_Node_Label(t *testing.T) {
	bgCtx := context.Background()
	filters := []bson.M{
		{},
		{"type": entity.NodeLabelSpec},
		{"type": entity.NodeLabelCPU},
		{"type": entity.NodeLabelMemory},
		{"type": entity.NodeLabelExclusiveTypeDeployment},
		{"type": entity.NodeLabelExclusiveTypeJob},
	}

	for i := range filters {
		filter := filters[i]

		typ := entity.NodeLabelKeyType("no type")
		if len(filter) > 0 {
			typ = filter["type"].(entity.NodeLabelKeyType)
		}
		t.Run(fmt.Sprintf("test for type=%s", typ), func(t *testing.T) {
			labels, err := d.FindNodeLabelLists(bgCtx, filter)
			assert.NoError(t, err)
			assert.NotEmpty(t, labels)
			if len(filter) > 0 {
				assert.Equal(t, 1, len(labels))
				assert.Equal(t, typ, labels[0].Type)
			}

			for _, label := range labels {
				switch label.Type {
				case entity.NodeLabelSpec:
					for _, val := range label.Values {
						switch val {
						case string(entity.ApplicationImportanceTypeLow),
							string(entity.ApplicationImportanceTypeMedium),
							string(entity.ApplicationImportanceTypeHigh),
							string(entity.ApplicationImportanceTypeSpecial):

						default:
							assert.Fail(
								t,
								"Invalid node_label value", "type=%s, value=%s",
								label.Type, val)
						}
					}

				case entity.NodeLabelCPU:

				case entity.NodeLabelMemory:

				case entity.NodeLabelExclusiveTypeDeployment:

				case entity.NodeLabelExclusiveTypeJob:

				default:
					assert.Fail(t, "Invalid node_label type", "type=%s", label.Type)
				}
			}
		})
	}
}
