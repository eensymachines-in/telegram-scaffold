package send

/*
The Finder interface is used by distributors getting the relevant groups from the BotResponse.
Selection of groups is based on certain criteria, in this case, the category of the message.
The dbFinder struct implements the Finder interface and uses the gorm.DB connection to get the relevant groups.
*/
import (
	"fmt"

	"github.com/eensymachines.in/telegram-scaffold/models"
	"gorm.io/gorm"
)

type Finder interface {
	FindDistributionIds(models.BotResponse) ([]int64, error)
}

var (
	// NewDBFinder : NewDBFinder is a function that returns a new dbFinder given a new DB connection
	NewDBFinder = func(db *gorm.DB) Finder {
		return &dbFinder{DB: db}
	}
)

// dbFinder : distributor functions need agents that can find relevant chatids to fwd to
// this one will, depending on the message category get the chatids for the botresponse from the DB connection
type dbFinder struct {
	*gorm.DB
}

// FindDistributionIds: FindDistributionIds gets the chatids for the botresponse from the DB connection
// the chatids are selected based on the category of the message
func (df *dbFinder) FindDistributionIds(res models.BotResponse) ([]int64, error) {
	chatIds := []int64{}
	result := []*models.TelegGrp{}
	catg := res.(models.CategorisedBotResponse).GetCategory()
	tx := df.Model(&models.TelegGrp{}).Where("? = ANY(categories)", catg).Find(&result) // picking the correct set of groups from the database
	if tx.Error != nil {
		return chatIds, fmt.Errorf("❗failed to retrieve the groups, %s", tx.Error) //failed query to get groups
	}
	if len(result) == 0 {
		return chatIds, fmt.Errorf("❗No relevant groups found for the category")
	}
	for _, grp := range result {
		chatIds = append(chatIds, grp.ChatID)
	}
	return chatIds, nil
}

// ----- Extend Finder interface for more implementations like dbFinder
