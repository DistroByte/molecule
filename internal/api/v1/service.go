package v1

type MoleculeAPIService struct {
	nomadService NomadServiceInterface
}

func NewMoleculeAPIService(nomadService NomadServiceInterface) *MoleculeAPIService {
	return &MoleculeAPIService{nomadService: nomadService}
}
