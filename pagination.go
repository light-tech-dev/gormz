package gormz

// ═══════════════════════════════════════════════
// Pagination
// ═══════════════════════════════════════════════

// PaginatedResult هو ناتج QuerySet.Paginate().
type PaginatedResult[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// IsEmpty يفحص إذا كانت النتيجة فارغة.
func (p *PaginatedResult[T]) IsEmpty() bool {
	return len(p.Items) == 0
}

// Len يرجّع عدد العناصر في الصفحة.
func (p *PaginatedResult[T]) Len() int {
	return len(p.Items)
}

// First يرجّع أول عنصر و ok.
//
//	first, ok := page.First()
//	if ok {
//	    fmt.Println(first.Name)
//	}
func (p *PaginatedResult[T]) First() (T, bool) {
	var zero T
	if len(p.Items) == 0 {
		return zero, false
	}
	return p.Items[0], true
}

// Last يرجّع آخر عنصر و ok.
func (p *PaginatedResult[T]) Last() (T, bool) {
	var zero T
	if len(p.Items) == 0 {
		return zero, false
	}
	return p.Items[len(p.Items)-1], true
}

// ForEach يتنقل على كل العناصر.
//
//	page.ForEach(func(i int, u User) {
//	    fmt.Println(i, u.Name)
//	})
func (p *PaginatedResult[T]) ForEach(fn func(int, T)) {
	if fn == nil {
		return
	}
	for i, item := range p.Items {
		fn(i, item)
	}
}

// Map يحوّل العناصر.
//
//	names := MapPage(page, func(u User) string { return u.Name })
func MapPage[T, U any](p *PaginatedResult[T], fn func(T) U) []U {
	if fn == nil {
		return nil
	}
	out := make([]U, len(p.Items))
	for i, item := range p.Items {
		out[i] = fn(item)
	}
	return out
}

// FilterPage يرجع عناصر مطابقة.
func FilterPage[T any](p *PaginatedResult[T], fn func(T) bool) []T {
	if fn == nil {
		return nil
	}
	var out []T
	for _, item := range p.Items {
		if fn(item) {
			out = append(out, item)
		}
	}
	return out
}

// ═══════════════════════════════════════════════
// Pagination Helper
// ═══════════════════════════════════════════════

// Page يمثل معاملات صفحة.
type Page struct {
	Number  int `json:"page"`
	PerPage int `json:"per_page"`
}

// NewPage ينشئ Page مع قيم افتراضية.
func NewPage(number, perPage int) Page {
	if number < 1 {
		number = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 1000 {
		perPage = 1000
	}
	return Page{Number: number, PerPage: perPage}
}

// Offset يرجّع offset الصفحة.
func (p Page) Offset() int {
	return (p.Number - 1) * p.PerPage
}