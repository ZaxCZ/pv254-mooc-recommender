package internal

import (
	"math"
)

// Course ...
type Course struct {
	ID              string   `json:"id" bson:"_id"`
	Categories      []string `json:"categories" bson:"categories"`
	Description     string   `json:"description" bson:"description"`
	Details         Details  `json:"details" bson:"details"`
	InterestedCount int32    `json:"interested_count" bson:"interested_count"`
	Link            string   `json:"link" bson:"link"`
	Name            string   `json:"name" bson:"name"`
	Overview        string   `json:"overview" bson:"overview"`
	Provider        string   `json:"provider" bson:"provider"`
	Rating          *float64 `json:"rating" bson:"rating"`
	ReviewCount     int32    `json:"review_count" bson:"review_count"`
	Schools         []string `json:"schools" bson:"schools"`
	Subject         string   `json:"subject" bson:"subject"`
	Syllabus        *string  `json:"syllabus" bson:"syllabus"`
	Teachers        []string `json:"teachers" bson:"teachers"`
}

// Details ...
type Details struct {
	Certificate      string   `json:"certificate" bson:"certificate"`
	Cost             int32    `json:"cost" bson:"cost"`
	Currency         string   `json:"currency" bson:"currency"`
	Duration         *float64 `json:"duration" bson:"duration"`
	DurationTimeUnit string   `json:"duration_time_unit" bson:"duration_time_unit"`
	Effort           *float64 `json:"effort" bson:"effort"`
	EffortTimeUnit   string   `json:"effort_time_unit" bson:"effort_time_unit"`
	Language         string   `json:"language" bson:"language"`
	Provider         string   `json:"provider" bson:"provider"`
	Session          string   `json:"session" bson:"session"`
	StartDate        []string `json:"start_date" bson:"start_date"`
}

//SortedBySimilarity ...
type SortedBySimilarity struct {
	coursesWithSimilarity []SimilarCourse
	course                *Course
}

//SimilarCourse ...
type SimilarCourse struct {
	Course     Course
	Similarity float64
}

func (s SortedBySimilarity) Len() int { return len(s.coursesWithSimilarity) }
func (s SortedBySimilarity) Swap(i, j int) {
	s.coursesWithSimilarity[i], s.coursesWithSimilarity[j] = s.coursesWithSimilarity[j], s.coursesWithSimilarity[i]
}
func (s SortedBySimilarity) Less(i, j int) bool {
	return s.coursesWithSimilarity[i].Similarity < s.coursesWithSimilarity[j].Similarity
}

//CourseSimVal ...
type CourseSimVal struct {
	Course *Course
	SimVal float64
}

//FindSimilar ...
func (c *Course) FindSimilar(courses []Course, similarityThreshold float64) []SimilarCourse {
	var result []SimilarCourse
	var courseOverviews []string
	for i := range courses {
		courseOverviews = append(courseOverviews, courses[i].Overview)
	}
	idf := computeIdf(courseOverviews)

	var vectDists []CourseSimVal
	for i := range courses {
		simVal := c.isSimilar(&courses[i], &idf)
		vectDists = append(vectDists, CourseSimVal{&courses[i], simVal})
	}

	maxVal := 0.0
	var normalizedDists []CourseSimVal
	for i := range vectDists {
		if vectDists[i].SimVal > 3 {
			continue
		}
		if maxVal < vectDists[i].SimVal {
			maxVal = vectDists[i].SimVal
		}
		normalizedDists = append(normalizedDists, vectDists[i])
	}

	for i := range normalizedDists {
		normalizedDists[i].SimVal = math.Abs(normalizedDists[i].SimVal/maxVal - 1)
	}

	for i := range normalizedDists {
		simVal := normalizedDists[i].SimVal
		if simVal > similarityThreshold {
			result = append(result, SimilarCourse{Course: *(normalizedDists[i].Course), Similarity: simVal})
		}
	}
	return result
}

func (c *Course) tfidf(idf map[string]float64) *map[string]float64 {
	tfidf := make(map[string]float64)
	tf := computeTf(c.Overview)
	for word, val := range *tf {
		tfidf[word] = val * idf[word]
	}
	return &tfidf
}

func (c *Course) isSimilar(c1 *Course, idf *map[string]float64) float64 {
	if c.ID == c1.ID {
		return 1.0
	}
	tfidf1 := c.tfidf(*idf)
	tfidf2 := c1.tfidf(*idf)

	res := 0.0
	wordList := make(map[string]bool)

	if len(*tfidf1) == 0 || len(*tfidf2) == 0 {
		return 3 // ja neviem kolko mam returnut ked jeden je uplne odveci
	}
	for word := range *tfidf1 {
		wordList[word] = true
	}
	for word := range *tfidf2 {
		wordList[word] = true
	}
	for word := range wordList {
		val1, val2 := 0.0, 0.0
		if num, ok := (*tfidf1)[word]; ok {
			val1 = num
		}
		if num, ok := (*tfidf2)[word]; ok {
			val2 = num
		}
		res += math.Pow(val1-val2, 2)
	}
	return math.Sqrt(res)
}
