package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"

	"github.com/shopspring/decimal"
)

type auctionService struct {
	auctionRepo repository.IAuctionRepository
}

func InitAuctionService(auctionRepo repository.IAuctionRepository) IAuctionService {
	return &auctionService{
		auctionRepo: auctionRepo,
	}
}

type IAuctionService interface {
	GetAuctionDataForPublic(auctionID uint) ([]dto.AuctionDataForPublic, error)
	GetAuctionDataForPrivate(auctionID, userID uint) ([]dto.AuctionDataForPrivate, error)
	SaveParticipantChanges(userID uint, participantData []dto.ParticipantDataForSave) error
}

func (service *auctionService) GetAuctionDataForPublic(auctionID uint) ([]dto.AuctionDataForPublic, error) {
	publicAuctionDataRaw, err := service.auctionRepo.GetAuctionDataForPublic(auctionID)
	if err != nil {
		return []dto.AuctionDataForPublic{}, nil
	}

	result := []dto.AuctionDataForPublic{}
	for index, raw := range publicAuctionDataRaw {
		totalPriceDecimal := raw.ParticipantPrice.Mul(decimal.NewFromFloat(raw.ItemQuantity))
		totalPrice, exact := totalPriceDecimal.Float64()
		if !exact {
			return []dto.AuctionDataForPublic{}, err
		}

		if index == 0 {
			entry := dto.AuctionDataForPublic{
				PackageName: raw.PackageName,
				PackageItems: []dto.AuctionItemDataForPublic{{
					Name:        raw.ItemName,
					Description: raw.ItemDescription,
					Unit:        raw.ItemUnit,
					Quantity:    raw.ItemQuantity,
					Note:        raw.ItemNote,
				}},
				ParticipantTotalPriceForPackage: []dto.AuctionParticipantTotalForPackageForPublic{},
			}
			if raw.ParticipantTitle != "" && totalPrice != 0 {
				entry.ParticipantTotalPriceForPackage = append(entry.ParticipantTotalPriceForPackage, dto.AuctionParticipantTotalForPackageForPublic{
					ParticipantTitle: raw.ParticipantTitle,
					TotalPrice:       totalPrice,
				})
			}

			result = append(result, entry)
		}

		lastPackageIndex := len(result) - 1
		if raw.PackageName == result[lastPackageIndex].PackageName {
			itemExists := false
			for _, item := range result[lastPackageIndex].PackageItems {
				if raw.ItemName == item.Name {
					itemExists = true
					break
				}
			}
			if !itemExists {
				result[lastPackageIndex].PackageItems = append(result[lastPackageIndex].PackageItems, dto.AuctionItemDataForPublic{
					Name:        raw.ItemName,
					Description: raw.ItemDescription,
					Unit:        raw.ItemUnit,
					Quantity:    raw.ItemQuantity,
					Note:        raw.ItemNote,
				})
			}

			participantIndex := -1
			for index, participant := range result[lastPackageIndex].ParticipantTotalPriceForPackage {
				if raw.ParticipantTitle == participant.ParticipantTitle {
					participantIndex = index
					break
				}
			}

			if participantIndex != -1 {
				result[lastPackageIndex].ParticipantTotalPriceForPackage[participantIndex].TotalPrice += totalPrice
			} else {
				result[lastPackageIndex].ParticipantTotalPriceForPackage = append(result[lastPackageIndex].ParticipantTotalPriceForPackage, dto.AuctionParticipantTotalForPackageForPublic{
					ParticipantTitle: raw.ParticipantTitle,
					TotalPrice:       totalPrice,
				})
			}
		} else {
			entry := dto.AuctionDataForPublic{
				PackageName: raw.PackageName,
				PackageItems: []dto.AuctionItemDataForPublic{{
					Name:        raw.ItemName,
					Description: raw.ItemDescription,
					Unit:        raw.ItemUnit,
					Quantity:    raw.ItemQuantity,
					Note:        raw.ItemNote,
				}},
				ParticipantTotalPriceForPackage: []dto.AuctionParticipantTotalForPackageForPublic{},
			}
			if raw.ParticipantTitle != "" && totalPrice != 0 {
				entry.ParticipantTotalPriceForPackage = append(entry.ParticipantTotalPriceForPackage, dto.AuctionParticipantTotalForPackageForPublic{
					ParticipantTitle: raw.ParticipantTitle,
					TotalPrice:       totalPrice,
				})
			}

			result = append(result, entry)
		}
	}

	return result, nil
}

func (service *auctionService) GetAuctionDataForPrivate(auctionID, userID uint) ([]dto.AuctionDataForPrivate, error) {
	privateDataForAuction, err := service.auctionRepo.GetAuctionDataForPrivate(auctionID)
	if err != nil {
		return []dto.AuctionDataForPrivate{}, err
	}

	result := []dto.AuctionDataForPrivate{}
	for index, raw := range privateDataForAuction {
		totalPriceDecimal := raw.ParticipantPrice.Mul(decimal.NewFromFloat(raw.ItemQuantity))
		totalPrice, exact := totalPriceDecimal.Float64()
		if !exact {
			return []dto.AuctionDataForPrivate{}, err
		}

		if index == 0 {
			entry := dto.AuctionDataForPrivate{
				PackageName: raw.PackageName,
				PackageItems: []dto.AuctionItemDataForPrivate{{
					ID:          raw.ItemID,
					Name:        raw.ItemName,
					Description: raw.ItemDescription,
					Unit:        raw.ItemUnit,
					Quantity:    raw.ItemQuantity,
					Note:        raw.ItemNote,
				}},
				ParticipantTotalPriceForPackage: []dto.AuctionParticipantTotalForPackageForPrivate{},
			}
			if raw.ParticipantUserID == userID {
				pricePerUnitFloat, exact := raw.ParticipantPrice.Float64()
				if !exact {
					return []dto.AuctionDataForPrivate{}, err
				}
				entry.PackageItems[0].UserUnitPrice = pricePerUnitFloat
				entry.PackageItems[0].Comment = raw.ParticipantComment
			}
			if raw.ParticipantTitle != "" && totalPrice != 0 {
				entry.ParticipantTotalPriceForPackage = append(entry.ParticipantTotalPriceForPackage, dto.AuctionParticipantTotalForPackageForPrivate{
					ParticipantTitle: raw.ParticipantTitle,
					TotalPrice:       totalPrice,
					IsCurrentUser:    false,
				})

				if raw.ParticipantUserID == userID {
					entry.ParticipantTotalPriceForPackage[0].IsCurrentUser = true
				}
			}

			result = append(result, entry)
      continue
		}

		lastPackageIndex := len(result) - 1
		if raw.PackageName == result[lastPackageIndex].PackageName {
			itemIndex := -1
			for subIndex, item := range result[lastPackageIndex].PackageItems {
				if raw.ItemName == item.Name {
					itemIndex = subIndex
					break
				}
			}
			if itemIndex == -1 {
				result[lastPackageIndex].PackageItems = append(result[lastPackageIndex].PackageItems, dto.AuctionItemDataForPrivate{
					ID:          raw.ItemID,
					Name:        raw.ItemName,
					Description: raw.ItemDescription,
					Unit:        raw.ItemUnit,
					Quantity:    raw.ItemQuantity,
					Note:        raw.ItemNote,
				})
        
        itemIndex = len(result[lastPackageIndex].PackageItems) - 1
			}

			if raw.ParticipantUserID == userID {
				pricePerUnitFloat, exact := raw.ParticipantPrice.Float64()
				if !exact {
					return []dto.AuctionDataForPrivate{}, err
				}

				result[lastPackageIndex].PackageItems[itemIndex].UserUnitPrice = pricePerUnitFloat
				result[lastPackageIndex].PackageItems[itemIndex].Comment = raw.ParticipantComment
			}

			participantIndex := -1
			for subIndex, participant := range result[lastPackageIndex].ParticipantTotalPriceForPackage {
				if raw.ParticipantTitle == participant.ParticipantTitle {
					participantIndex = subIndex
					break
				}
			}

			if participantIndex != -1 {
				result[lastPackageIndex].ParticipantTotalPriceForPackage[participantIndex].TotalPrice += totalPrice
			} else {
				result[lastPackageIndex].ParticipantTotalPriceForPackage = append(result[lastPackageIndex].ParticipantTotalPriceForPackage, dto.AuctionParticipantTotalForPackageForPrivate{
					ParticipantTitle: raw.ParticipantTitle,
					TotalPrice:       totalPrice,
					IsCurrentUser:    raw.ParticipantUserID == userID,
				})
			}
		} else {
			entry := dto.AuctionDataForPrivate{
				PackageName: raw.PackageName,
				PackageItems: []dto.AuctionItemDataForPrivate{{
					ID:          raw.ItemID,
					Name:        raw.ItemName,
					Description: raw.ItemDescription,
					Unit:        raw.ItemUnit,
					Quantity:    raw.ItemQuantity,
					Note:        raw.ItemNote,
				}},
				ParticipantTotalPriceForPackage: []dto.AuctionParticipantTotalForPackageForPrivate{},
			}
			if raw.ParticipantUserID == userID {
				pricePerUnitFloat, exact := raw.ParticipantPrice.Float64()
				if !exact {
					return []dto.AuctionDataForPrivate{}, err
				}
				entry.PackageItems[0].UserUnitPrice = pricePerUnitFloat
				entry.PackageItems[0].Comment = raw.ParticipantComment
			}

			if raw.ParticipantTitle != "" && totalPrice != 0 {
				entry.ParticipantTotalPriceForPackage = append(entry.ParticipantTotalPriceForPackage, dto.AuctionParticipantTotalForPackageForPrivate{
					ParticipantTitle: raw.ParticipantTitle,
					TotalPrice:       totalPrice,
					IsCurrentUser:    false,
				})

				if raw.ParticipantUserID == userID {
					entry.ParticipantTotalPriceForPackage[0].IsCurrentUser = true
				}
			}

			result = append(result, entry)
		}
	}

  for index, auctionPackage := range result {
    for _, totalPrice := range auctionPackage.ParticipantTotalPriceForPackage {
      if result[index].MinimumPackagePrice == 0 {
        result[index].MinimumPackagePrice = totalPrice.TotalPrice
        continue
      } 

      if result[index].MinimumPackagePrice > totalPrice.TotalPrice {
        result[index].MinimumPackagePrice = totalPrice.TotalPrice
      }
    }
  }

	return result, nil
}

func (service *auctionService) SaveParticipantChanges(userID uint, participantData []dto.ParticipantDataForSave) error {
	auctionParticipantPrice := []model.AuctionParticipantPrice{}
	for _, entry := range participantData {
		auctionParticipantPrice = append(auctionParticipantPrice, model.AuctionParticipantPrice{
			AuctionItemID: entry.ItemID,
			UserID:        userID,
			UnitPrice:     entry.UnitPrice,
			Comments:      entry.Comment,
		})
	}

	return service.auctionRepo.SaveParticipantChanges(auctionParticipantPrice)
}
