package constants

// French strings
const (
	// Account related
	CreateAccount = "créer un compte"
	Login         = "se connecter"
	Connection    = "connexion"

	// Registration related
	SignUp              = "s'inscrire"
	Register            = "inscription"
	Registers           = "inscriptions"
	ToRegister          = "inscrire"
	SignUpForm          = "formulaire"
	Participate         = "participer"
	Registration        = "enregistrer"
	ClosedRegistration  = "inscription fermée"
	ClosedRegistrations = "inscriptions fermées"
	RegistrationsClosed = "inscriptions terminées"
	RegistrationClosed  = "inscription terminée"

	// Navigation related
	NextStep         = "étape suivante"
	NextStepNoAccent = "etape suivante"
	Next             = "suivant"
	Continue         = "continuer"

	// Payment related
	Payment  = "paiement"
	Pay      = "payer"
	Pricing  = "tarif"
	Pricings = "tarifs"

	// Engagement related
	Engagement  = "engagement"
	Engagements = "engagements"

	// Online related
	Online = "en ligne"
	On     = "sur"
)

// English strings
const (
	// Account related
	ENCreateAccount = "create account"
	ENLogin         = "login"
	ENSignIn        = "sign in"

	// Registration related
	ENSignUp       = "signup"
	ENSignUpHyphen = "sign-up"
	ENSignUpSpace  = "sign up"
	ENRegister     = "register"
	ENRegistration = "registration"
	ENParticipate  = "participate"

	// Navigation related
	ENNextStep = "next step"
	ENContinue = "continue"
)

// Grouped keywords for common searches
var (
	// All registration-related keywords in French and English
	RegistrationKeywords = []string{
		Register, Registers, ToRegister, SignUp,
		ENRegister, ENRegistration, ENSignUp, ENSignUpHyphen, ENSignUpSpace,
		"registre", "s'enregistrer",
		SignUpForm, Participate, Registration,
		Engagement, Engagements,
		NextStep, NextStepNoAccent, Next, Continue,
	}

	// Signup-specific keywords
	SignupKeywords = []string{
		Register, ENRegister, ENSignUp, SignUp, SignUpForm, Participate, Registration,
	}

	// High-value keywords that strongly indicate signup functionality
	HighValueKeywords = []string{
		Register, ENSignUp, ENRegister, SignUp,
	}
)
