package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type auctionRepository struct {
	db *gorm.DB
}

func InitAuctionRepository(db *gorm.DB) IAuctionRepository {
	return &auctionRepository{
		db: db,
	}
}

type IAuctionRepository interface {
	GetAuctionDataForPublic(auctionID uint) ([]dto.AuctionDataForPublicQueryResult, error)
	GetAuctionDataForPrivate(auctionID uint) ([]dto.AuctionDataForPrivateQueryResult, error)
	SaveParticipantChanges(auctionParticipantPrice []model.AuctionParticipantPrice) error
}

func (repo *auctionRepository) GetAuctionDataForPublic(auctionID uint) ([]dto.AuctionDataForPublicQueryResult, error) {
	result := []dto.AuctionDataForPublicQueryResult{}
	err := repo.db.Raw(`
    SELECT 
      auction_packages.id as package_id,
      auction_packages.name as package_name,
      auction_items.name as item_name,
      auction_items.description as item_description,
      auction_items.unit as item_unit,
      auction_items.quantity as item_quantity,
      auction_items.note as item_note,
      auction_participant_prices.unit_price as participant_price,
      workers.job_title_in_project as participant_title
    FROM auction_items
    FULL JOIN auction_packages ON auction_packages.id = auction_items.auction_package_id
    FULL JOIN auctions ON auctions.id = auction_packages.auction_id
    LEFT JOIN auction_participant_prices ON auction_participant_prices.auction_item_id = auction_items.id
    LEFT JOIN users ON users.id = auction_participant_prices.user_id
    LEFT JOIN workers ON workers.id = users.worker_id 
    WHERE auctions.id = ?
    ORDER BY auction_packages.id, auction_items.id, auction_participant_prices.user_id
    `, auctionID).Scan(&result).Error

	return result, err
}

func (repo *auctionRepository) GetAuctionDataForPrivate(auctionID uint) ([]dto.AuctionDataForPrivateQueryResult, error) {
	result := []dto.AuctionDataForPrivateQueryResult{}
	err := repo.db.Raw(`
      SELECT 
        auction_packages.id as package_id,
        auction_packages.name as package_name,
        auction_items.id as item_id,
        auction_items.name as item_name,
        auction_items.description as item_description,
        auction_items.unit as item_unit,
        auction_items.quantity as item_quantity,
        auction_items.note as item_note,
        auction_participant_prices.comments as participant_comment,
        auction_participant_prices.unit_price as participant_price,
        users.id as participant_user_id,
        workers.job_title_in_project as participant_title
      FROM auction_items
      FULL JOIN auction_packages ON auction_packages.id = auction_items.auction_package_id
      FULL JOIN auctions ON auctions.id = auction_packages.auction_id
      LEFT JOIN auction_participant_prices ON auction_participant_prices.auction_item_id = auction_items.id
      LEFT JOIN users ON users.id = auction_participant_prices.user_id
      LEFT JOIN workers ON workers.id = users.worker_id 
      WHERE auctions.id = ?
      ORDER BY auction_packages.id
    `, auctionID).Scan(&result).Error
	return result, err
}

func(repo *auctionRepository)SaveParticipantChanges(auctionParticipantPrice []model.AuctionParticipantPrice) error {
  return repo.db.Transaction(func(tx *gorm.DB) error {
    for _, entry := range auctionParticipantPrice {
      err := repo.db.
        Where(model.AuctionParticipantPrice{
          AuctionItemID: entry.AuctionItemID,
          UserID: entry.UserID,
        }).
        Assign(model.AuctionParticipantPrice{
          UnitPrice: entry.UnitPrice,
          Comments: entry.Comments,
        }).
        FirstOrCreate(&entry).
        Error
      
      if err != nil {
        return err
      }
    }

    return nil
  })
}
