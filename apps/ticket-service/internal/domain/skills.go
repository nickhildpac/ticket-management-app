package domain

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
)

// ValidSkillsList that can be assigned to users and tickets
var ValidSkillsList = []string{
	"incident-management",
	"major-incident",
	"root-cause-analysis",
	"log-analysis",
	"production-support",
	"sla-management",
	"post-incident-review",
}

// Skills represents a collection of skills with validation
type Skills struct {
	items []string
}

// NewSkills creates a new Skills instance with validation
func NewSkills(skills []string) (*Skills, error) {
	s := &Skills{items: make([]string, 0, len(skills))}
	for _, skill := range skills {
		trimmed := strings.TrimSpace(strings.ToLower(skill))
		if trimmed == "" {
			continue // skip empty skills
		}
		if !IsValidSkill(trimmed) {
			return nil, errors.New("invalid skill: " + skill)
		}
		// Avoid duplicates
		if !s.Contains(trimmed) {
			s.items = append(s.items, trimmed)
		}
	}
	return s, nil
}

// MustNewSkills creates Skills and panics if validation fails
// Useful for tests or known-good data
func MustNewSkills(skills []string) Skills {
	s, err := NewSkills(skills)
	if err != nil {
		panic(err)
	}
	return *s
}

// NewSkillsFromSlice creates Skills from a string slice without validation
// This is used when loading from the database where skills are already validated
func NewSkillsFromSlice(skills []string) Skills {
	if skills == nil {
		return Skills{items: []string{}}
	}
	return Skills{items: skills}
}

// ToSlice returns the underlying slice
func (s *Skills) ToSlice() []string {
	if s == nil {
		return nil
	}
	return s.items
}

// Contains checks if a skill exists in the collection
func (s *Skills) Contains(skill string) bool {
	if s == nil {
		return false
	}
	return slices.Contains(s.items, skill)
}

// Add adds a new skill with validation
func (s *Skills) Add(skill string) error {
	trimmed := strings.TrimSpace(strings.ToLower(skill))
	if trimmed == "" {
		return errors.New("skill cannot be empty")
	}
	if !IsValidSkill(trimmed) {
		return errors.New("invalid skill: " + skill)
	}
	if !s.Contains(trimmed) {
		s.items = append(s.items, trimmed)
	}
	return nil
}

// Remove removes a skill from the collection
func (s *Skills) Remove(skill string) {
	trimmed := strings.TrimSpace(strings.ToLower(skill))
	for i, item := range s.items {
		if item == trimmed {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return
		}
	}
}

// Size returns the number of skills
func (s *Skills) Size() int {
	if s == nil {
		return 0
	}
	return len(s.items)
}

// IsValidSkill checks if a skill is in the valid skills list
func IsValidSkill(skill string) bool {
	skill = strings.ToLower(strings.TrimSpace(skill))
	return slices.Contains(ValidSkillsList, skill)
}

// GetValidSkills returns the list of all valid skills
func GetValidSkills() []string {
	return append([]string{}, ValidSkillsList...)
}

// MarshalJSON implements json.Marshaler for Skills
func (s Skills) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.items)
}

// UnmarshalJSON implements json.Unmarshaler for Skills
func (s *Skills) UnmarshalJSON(data []byte) error {
	var slice []string
	if err := json.Unmarshal(data, &slice); err != nil {
		return err
	}
	skills, err := NewSkills(slice)
	if err != nil {
		return err
	}
	s.items = skills.items
	return nil
}
