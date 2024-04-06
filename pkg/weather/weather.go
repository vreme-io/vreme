package weather

// Common weather provider interface
type WeatherProvider interface {
	Update() (map[string]string, map[string]string, error)
	GetMETAR(station string) (string, error)
	GetTAF(station string) (string, error)
}

type WeatherService struct {
	provider WeatherProvider
}

func NewService(provider WeatherProvider) *WeatherService {
	return &WeatherService{provider}
}

func (s *WeatherService) Update() (map[string]string, map[string]string, error) {
	return s.provider.Update()
}

func (s *WeatherService) GetMETAR(station string) (string, error) {
	return s.provider.GetMETAR(station)
}

func (s *WeatherService) GetTAF(station string) (string, error) {
	return s.provider.GetTAF(station)
}
