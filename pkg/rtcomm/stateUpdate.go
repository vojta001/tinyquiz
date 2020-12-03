package rtcomm

type StateUpdate struct {
	Players []Player `json:"players"`
	Question *QuestionUpdate `json:"question,omitempty"`
}

type Player struct {
	Organiser bool `json:"organiser"`
	Name string `json:"name"`
}

type QuestionUpdate struct {
	Title string `json:"title"`
	Answers []Answer `json:"answers"`
}

type Answer struct{
	ID string `json:"id"`
	Title string `json:"title"`
}
