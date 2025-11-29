package product

// ProductRepository defines the product storage API used by service layer.
type ProductRepository interface {
	Create(*Product) error
	GetByID(string) (*Product, error)
	List() ([]Product, error)
	Update(*Product) error
	SoftDelete(id, deletedBy string) error
}

type Service struct{ repo ProductRepository }

func NewService(r ProductRepository) *Service { return &Service{repo: r} }

func (s *Service) Create(p *Product) error         { return s.repo.Create(p) }
func (s *Service) Get(id string) (*Product, error) { return s.repo.GetByID(id) }
func (s *Service) List() ([]Product, error)        { return s.repo.List() }
func (s *Service) Update(p *Product) error         { return s.repo.Update(p) }
func (s *Service) Delete(id, by string) error      { return s.repo.SoftDelete(id, by) }
