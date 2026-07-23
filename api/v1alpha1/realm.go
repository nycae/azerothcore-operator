package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServerType string

const (
	ServerTypePvP   ServerType = "pvp"
	ServerTypePvE   ServerType = "pve"
	ServerTypeRp    ServerType = "rp"
	ServerTypeRpPvP ServerType = "rppvp"
)

func (s ServerType) ToInt() int8 {
	switch s {
	case ServerTypePvP:
		return 1
	case ServerTypeRp:
		return 6
	case ServerTypeRpPvP:
		return 8
	case ServerTypePvE:
		fallthrough
	default:
		return 4
	}
}

type RealmFlag string

const (
	RealmFlagNone         RealmFlag = "None"
	RealmFlagInvalid      RealmFlag = "Invalid"
	RealmFlagOffline      RealmFlag = "Offline"
	RealmFlagSpecifyBuild RealmFlag = "SpecifyBuild"
	RealmFlagMedium       RealmFlag = "Medium"
	RealmFlagNewPlayers   RealmFlag = "New Players"
	RealmFlagRecommended  RealmFlag = "Recommended"
	RealmFlagFull         RealmFlag = "Full"
)

func (s RealmFlag) ToInt() int8 {
	switch s {
	case RealmFlagInvalid:
		return 1
	case RealmFlagOffline:
		return 2
	case RealmFlagSpecifyBuild:
		return 4
	case RealmFlagMedium:
		return 8
	case RealmFlagNewPlayers:
		return 16
	case RealmFlagRecommended:
		return 32
	case RealmFlagFull:
		return 64
	case RealmFlagNone:
		fallthrough
	default:
		return 0
	}
}

type Timezone string

const (
	TimezoneDevelopment       Timezone = "Development"
	TimezoneUnitedStates      Timezone = "United States"
	TimezoneOceanic           Timezone = "Oceanic"
	TimezoneLatinAmerica      Timezone = "Latin America"
	TimezoneTournament        Timezone = "Tournament"
	TimezoneKorea             Timezone = "Korea"
	TimezoneEnglish           Timezone = "English"
	TimezoneGerman            Timezone = "German"
	TimezoneFrench            Timezone = "French"
	TimezoneSpanish           Timezone = "Spanish"
	TimezoneRussian           Timezone = "Russian"
	TimezoneTaiwan            Timezone = "Taiwan"
	TimezoneChina             Timezone = "China"
	TimezoneCN1               Timezone = "CN1"
	TimezoneCN2               Timezone = "CN2"
	TimezoneCN3               Timezone = "CN3"
	TimezoneCN4               Timezone = "CN4"
	TimezoneCN5               Timezone = "CN5"
	TimezoneCN6               Timezone = "CN6"
	TimezoneCN7               Timezone = "CN7"
	TimezoneCN8               Timezone = "CN8"
	TimezoneTestServer        Timezone = "Test Server"
	TimezoneCN9               Timezone = "CN9"
	TimezoneTestServer2       Timezone = "Test Server 2"
	TimezoneCN10              Timezone = "CN10"
	TimezoneCTC               Timezone = "CTC"
	TimezoneCNC               Timezone = "CNC"
	TimezoneCN1_4             Timezone = "CN1/4"
	TimezoneCN_2_6_9          Timezone = "CN/2/6/9"
	TimezoneCN3_7             Timezone = "CN3/7"
	TimezoneRussianTournament Timezone = "Russian Tournament"
	TimezoneCN5_8             Timezone = "CN5/8"
	TimezoneCN11              Timezone = "CN11"
	TimezoneCN12              Timezone = "CN12"
	TimezoneCN13              Timezone = "CN13"
	TimezoneCN14              Timezone = "CN14"
	TimezoneCN15              Timezone = "CN15"
	TimezoneCN16              Timezone = "CN16"
	TimezoneCN17              Timezone = "CN17"
	TimezoneCN18              Timezone = "CN18"
	TimezoneCN19              Timezone = "CN19"
	TimezoneCN20              Timezone = "CN20"
	TimezoneBrazil            Timezone = "Brazil"
	TimezoneItalian           Timezone = "Italian"
	TimezoneHyrule            Timezone = "Hyrule"
	TimezoneQA2Test           Timezone = "QA2 Test"
	TimezoneRecommendedRealm  Timezone = "Recommended Realm"
	TimezoneTest              Timezone = "Test"
	TimezoneFutureTest        Timezone = "Future Test"
)

func (s Timezone) ToInt() int8 {
	switch s {
	case TimezoneUnitedStates:
		return 2
	case TimezoneOceanic:
		return 3
	case TimezoneLatinAmerica:
		return 4
	case TimezoneTournament: // ID 5, 7, 13, 15, 25, 27
		return 5
	case TimezoneKorea:
		return 6
	case TimezoneEnglish:
		return 8
	case TimezoneGerman:
		return 9
	case TimezoneFrench:
		return 10
	case TimezoneSpanish:
		return 11
	case TimezoneRussian:
		return 12
	case TimezoneTaiwan:
		return 14
	case TimezoneChina:
		return 16
	case TimezoneCN1:
		return 17
	case TimezoneCN2:
		return 18
	case TimezoneCN3:
		return 19
	case TimezoneCN4:
		return 20
	case TimezoneCN5:
		return 21
	case TimezoneCN6:
		return 22
	case TimezoneCN7:
		return 23
	case TimezoneCN8:
		return 24
	case TimezoneTestServer:
		return 26
	case TimezoneCN9:
		return 29
	case TimezoneTestServer2:
		return 30
	case TimezoneCN10:
		return 31
	case TimezoneCTC:
		return 32
	case TimezoneCNC:
		return 33
	case TimezoneCN1_4:
		return 34
	case TimezoneCN_2_6_9:
		return 35
	case TimezoneCN3_7:
		return 36
	case TimezoneRussianTournament:
		return 37
	case TimezoneCN5_8:
		return 38
	case TimezoneCN11:
		return 39
	case TimezoneCN12:
		return 40
	case TimezoneCN13:
		return 41
	case TimezoneCN14:
		return 42
	case TimezoneCN15:
		return 43
	case TimezoneCN16:
		return 44
	case TimezoneCN17:
		return 45
	case TimezoneCN18:
		return 46
	case TimezoneCN19:
		return 47
	case TimezoneCN20:
		return 48
	case TimezoneBrazil:
		return 49
	case TimezoneItalian:
		return 50
	case TimezoneHyrule:
		return 51
	case TimezoneQA2Test:
		return 52
	case TimezoneRecommendedRealm: // ID 55, 57
		return 55
	case TimezoneTest:
		return 56
	case TimezoneFutureTest:
		return 59
	case TimezoneDevelopment:
		fallthrough
	default:
		return 1
	}
}

// Unused but cool
type GameBuild string

const (
	Patch1121 GameBuild = "1.12.1"
	Patch1122 GameBuild = "1.12.2"
	Patch243  GameBuild = "2.4.3"
	Patch313  GameBuild = "3.1.3"
	Patch320  GameBuild = "3.2.0"
	Patch322a GameBuild = "3.2.2a"
	Patch330  GameBuild = "3.3.0"
	Patch330a GameBuild = "3.3.0a"
	Patch332  GameBuild = "3.3.2"
	Patch333  GameBuild = "3.3.3"
	Patch333a GameBuild = "3.3.3a"
	Patch335a GameBuild = "3.3.5a"
)

func (p GameBuild) Build() uint32 {
	switch p {
	case Patch1121:
		return 5875
	case Patch1122:
		return 6005
	case Patch243:
		return 8606
	case Patch313:
		return 9947
	case Patch320:
		return 10146
	case Patch322a:
		return 10505
	case Patch330:
		return 10571
	case Patch330a:
		return 11159
	case Patch332:
		return 11403
	case Patch333:
		return 11623
	case Patch333a:
		return 11723
	case Patch335a:
		return 12340
	default:
		return 0
	}
}

type GmOnly bool

func (g GmOnly) ToInt() int8 {
	if g {
		return 1
	}
	return 0
}

type WorldServerSpec struct {
	MaxPlayers int32                       `json:"maxPlayers,omitempty"`
	Resources  corev1.ResourceRequirements `json:"resources,omitempty"`
}

type RealmAddressSpec struct {
	Address      string `json:"address"`
	LocalAddress string `json:"localAddress"`
	Port         uint16 `json:"port"`
}

type DatabaseStrategy string

const (
	DatabaseStrategySelfManaged DatabaseStrategy = "SelfManaged"
	DatabaseStrategyAutomatic   DatabaseStrategy = "Automatic"
)

type DatabaseConnection struct {
	Hostname          string            `json:"hostname"`
	Port              uint16            `json:"port"`
	Username          string            `json:"username"`
	PasswordSecretRef SecretKeySelector `json:"passwordSecretRef"`
	Database          string            `json:"database"`
}

type DatabaseSpec struct {
	Strategy    DatabaseStrategy    `json:"strategy"`
	WorldDB     *DatabaseConnection `json:"worldDB"`
	CharacterDB *DatabaseConnection `json:"characterDB"`
}

type RealmSpec struct {
	RealmName   string           `json:"name"`
	RealmType   ServerType       `json:"realmType"`
	Routing     RealmAddressSpec `json:"routing,omitempty"`
	Build       GameBuild        `json:"build"`
	GmOnly      GmOnly           `json:"gmOnly"`
	Status      RealmFlag        `json:"status,omitempty"`
	Timezone    Timezone         `json:"timezone,omitempty"`
	WorldServer WorldServerSpec  `json:"worldServer,omitempty"`
	Database    DatabaseSpec     `json:"database,omitempty"`
}

type RealmStatus struct {
	Ready   bool  `json:"ready,omitempty"`
	RealmID int64 `json:"realmId,omitempty"`
}

type Realm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RealmSpec   `json:"spec,omitempty"`
	Status RealmStatus `json:"status,omitempty"`
}

type RealmList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Realm `json:"items"`
}

func (in *Realm) DeepCopyObject() runtime.Object {
	out := *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec

	in.Spec.WorldServer.Resources.DeepCopyInto(&out.Spec.WorldServer.Resources)
	return &out
}

func (in *RealmList) DeepCopyObject() runtime.Object {
	out := *in
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]Realm, len(in.Items))
		for i := range in.Items {
			out.Items[i] = *in.Items[i].DeepCopyObject().(*Realm)
		}
	}
	return &out
}
