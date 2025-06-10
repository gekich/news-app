package functions

import (
	"fmt"
)

// toInt converts an interface to int safely
func toInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("value must be an integer")
	}
}

// PaginationFuncs returns pagination-related template functions
func PaginationFuncs() map[string]interface{} {
	return map[string]interface{}{
		"seq": func(start, end interface{}) ([]int, error) {
			startInt, err := toInt(start)
			if err != nil {
				return nil, fmt.Errorf("start %v", err)
			}

			endInt, err := toInt(end)
			if err != nil {
				return nil, fmt.Errorf("end %v", err)
			}

			seq := make([]int, 0, endInt-startInt+1)
			for i := startInt; i <= endInt; i++ {
				seq = append(seq, i)
			}
			return seq, nil
		},
		"pagesToShow": func(currentPage, totalPages interface{}) ([]int, error) {
			current, err := toInt(currentPage)
			if err != nil {
				return nil, fmt.Errorf("currentPage %v", err)
			}

			total, err := toInt(totalPages)
			if err != nil {
				return nil, fmt.Errorf("totalPages %v", err)
			}

			if total <= 7 {
				pages := make([]int, total)
				for i := 0; i < total; i++ {
					pages[i] = i + 1
				}
				return pages, nil
			}

			result := make([]int, 0, 7)

			result = append(result, 1)

			start := current - 2
			end := current + 2

			if start <= 1 {
				start = 2
				end = min(start+4, total-1)
			} else if end >= total {
				end = total - 1
				start = max(2, end-4)
			}

			if start > 2 {
				result = append(result, -1)
			}

			for i := start; i <= end; i++ {
				result = append(result, i)
			}

			if end < total-1 {
				result = append(result, -1)
			}

			result = append(result, total)

			return result, nil
		},
	}
}
