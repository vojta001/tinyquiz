package rtcomm

type StateUpdate struct {
	Players  []Player        `json:"players,omitempty"`
	Question *QuestionUpdate `json:"question,omitempty"`
	Results  bool            `json:"results,omitempty"`
}

type Player struct {
	Organiser bool   `json:"organiser"`
	Name      string `json:"name"`
}

type QuestionUpdate struct {
	Title         string   `json:"title"`
	RemainingTime uint64   `json:"remainingTime"`
	Answers       []Answer `json:"answers"`
}

type Answer struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
