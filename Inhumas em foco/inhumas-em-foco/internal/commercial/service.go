package commercial

import (
	"strings"
	"time"

	"inhumas-em-foco/internal/model"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func NormalizeBannerStatus(status string) string {
	clean := strings.ToLower(strings.TrimSpace(status))
	switch clean {
	case "draft", "paused", "expired":
		return clean
	default:
		return "active"
	}
}

func (s *Service) ValidateBanner(banner *model.Banner) string {
	if banner == nil {
		return "Banner invalido"
	}
	if strings.TrimSpace(banner.Name) == "" {
		return "Nome da campanha e obrigatorio"
	}
	if strings.TrimSpace(banner.AdvertiserName) == "" {
		return "Cliente/anunciante e obrigatorio"
	}
	if strings.TrimSpace(banner.Position) == "" {
		return "Posicao e obrigatoria"
	}
	if strings.TrimSpace(banner.LinkURL) == "" {
		return "Link de destino e obrigatorio"
	}
	if NormalizeBannerStatus(banner.Status) == "active" && strings.TrimSpace(banner.ImageKey) == "" {
		return "Banner ativo precisa de imagem"
	}
	return ""
}

func (s *Service) ParseDateRange(startValue, endValue string) (time.Time, time.Time, string) {
	start, err := time.Parse("2006-01-02", startValue)
	if err != nil {
		return time.Time{}, time.Time{}, "Data de inicio invalida"
	}
	end, err := time.Parse("2006-01-02", endValue)
	if err != nil {
		return time.Time{}, time.Time{}, "Data de fim invalida"
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, "Data de fim deve ser igual ou posterior a data de inicio"
	}
	return start, end, ""
}
