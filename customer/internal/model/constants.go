package model

// ContactType values for rel_contact.contact_type
const (
	ContactTypePhone   int16 = 1
	ContactTypeAddress int16 = 2
)

// LinkType values for rel_contact.link_type
const (
	LinkTypeIndividual  int16 = 1 // link_id = individual_customers.id
	LinkTypeBusiness    int16 = 2 // link_id = business_customers.id (company contacts)
	LinkTypeProprietor  int16 = 3 // link_id = business_customers.id (proprietor contacts)
)
