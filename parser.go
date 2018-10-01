package fullname_parser

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ParsedName struct {
	Title      string `json:"title,omitempty"`
	First      string `json:"first,omitempty"`
	Middle     string `json:"middle,omitempty"`
	Last       string `json:"last,omitempty"`
	Nick       string `json:"nick,omitempty"`
	Suffix     string `json:"suffix,omitempty"`
	nameParts  []string
	nameCommas []bool
	rawName    string
}

var (
	suffixList = []string{"esq", "esquire", "jr", "jnr", "sr", "snr", "2", "ii", "iii", "iv",
		"v", "clu", "chfc", "cfp", "md", "phd", "j.d.", "ll.m.", "m.d.", "d.o.", "d.c.",
		"p.c.", "ph.d."}

	prefixList = []string{"a", "ab", "antune", "ap", "abu", "al", "alm", "alt", "bab", "bäck",
		"bar", "bath", "bat", "beau", "beck", "ben", "berg", "bet", "bin", "bint", "birch",
		"björk", "björn", "bjur", "da", "dahl", "dal", "de", "degli", "dele", "del",
		"della", "der", "di", "dos", "du", "e", "ek", "el", "escob", "esch", "fleisch",
		"fitz", "fors", "gott", "griff", "haj", "haug", "holm", "ibn", "kauf", "kil",
		"koop", "kvarn", "la", "le", "lind", "lönn", "lund", "mac", "mhic", "mic", "mir",
		"na", "naka", "neder", "nic", "ni", "nin", "nord", "norr", "ny", "o", "ua", `ui\'`,
		"öfver", "ost", "över", "öz", "papa", "pour", "quarn", "skog", "skoog", "sten",
		"stor", "ström", "söder", "ter", "ter", "tre", "türk", "van", "väst", "väster",
		"vest", "von"}

	titleList = []string{"mr", "mrs", "ms", "miss", "dr", "herr", "monsieur", "hr", "frau",
		"a v m", "admiraal", "admiral", "air cdre", "air commodore", "air marshal",
		"air vice marshal", "alderman", "alhaji", "ambassador", "baron", "barones",
		"brig", "brig gen", "brig general", "brigadier", "brigadier general",
		"brother", "canon", "capt", "captain", "cardinal", "cdr", "chief", "cik", "cmdr",
		"coach", "col", "colonel", "commandant", "commander", "commissioner",
		"commodore", "comte", "comtessa", "congressman", "conseiller", "consul",
		"conte", "contessa", "corporal", "councillor", "count", "countess",
		"crown prince", "crown princess", "dame", "datin", "dato", "datuk",
		"datuk seri", "deacon", "deaconess", "dean", "dhr", "dipl ing", "doctor",
		"dott", "dott sa", "dr ing", "dra", "drs", "embajador", "embajadora", "en",
		"encik", "eng", "eur ing", "exma sra", "exmo sr", "f o", "father",
		"first lieutient", "first officer", "flt lieut", "flying officer", "fr",
		"frau", "fraulein", "fru", "gen", "generaal", "general", "governor", "graaf",
		"gravin", "group captain", "grp capt", "h e dr", "h h", "h m", "h r h", "hajah",
		"haji", "hajim", "her highness", "her majesty", "herr", "high chief",
		"his highness", "his holiness", "his majesty", "hon", "hr", "hra", "ing", "ir",
		"jonkheer", "judge", "justice", "khun ying", "kolonel", "lady", "lcda", "lic",
		"lieut", "lieut cdr", "lieut col", "lieut gen", "lord", "m", "m l", "m r",
		"madame", "mademoiselle", "maj gen", "major", "master", "mevrouw", "miss",
		"mlle", "mme", "monsieur", "monsignor", "mstr", "nti", "pastor",
		"president", "prince", "princess", "princesse", "prinses", "prof",
		"prof sir", "professor", "puan", "puan sri", "rabbi", "rear admiral", "rev",
		"rev canon", "rev dr", "rev mother", "reverend", "rva", "senator", "sergeant",
		"sheikh", "sheikha", "sig", "sig na", "sig ra", "sir", "sister", "sqn ldr", "sr",
		"sr d", "sra", "srta", "sultan", "tan sri", "tan sri dato", "tengku", "teuku",
		"than puying", "the hon dr", "the hon justice", "the hon miss", "the hon mr",
		"the hon mrs", "the hon ms", "the hon sir", "the very rev", "toh puan", "tun",
		"vice admiral", "viscount", "viscountess", "wg cdr"}

	conjunctionList = []string{"&", "and", "et", "e", "of", "the", "und", "y"}

	nickNameRE  = regexp.MustCompile(`\s?[\'\"\(\[]([^\[\]\)\)\'\"]+)[\'\"\)\]]`)
	splitNameRE = regexp.MustCompile(`[\s\p{Zs}]{2,}`)
)

func ParseFullname(fullname string) ParsedName {
	log.Debug("Start parsing fullname: ", fullname)

	parsedName := &ParsedName{
		rawName: fullname,
	}

	parsedName.Parse()

	return *parsedName
}

func (parsedName *ParsedName) Parse() {
	//nicknames: remove and store
	nicknames := parsedName.findNicknames()
	parsedName.Nick = strings.Join(nicknames, ",")

	//split name to parts and store commas
	parsedName.splitName()

	//suffix: remove and store
	if len(parsedName.nameParts) > 1 {
		suffixes := parsedName.findSuffixes()
		parsedName.Suffix = strings.Join(suffixes, ", ")
	}

	//titles: remove and store
	if len(parsedName.nameParts) > 1 {
		titles := parsedName.findTitles()
		parsedName.Title = strings.Join(titles, ", ")
	}
	log.Debugf("num parts: %d", len(parsedName.nameParts))

	// Join name prefixes to following names
	if len(parsedName.nameParts) > 1 {
		parsedName.joinPrefixes()
	}

	// Join conjunctions to surrounding names
	if len(parsedName.nameParts) > 1 {
		parsedName.joinConjunctions()
	}

	// Suffix: remove and store items after extra commas as suffixes
	if len(parsedName.nameParts) > 1 {
		extraSuffixes := parsedName.findExtraSuffixes()
		if len(extraSuffixes) > 0 {
			if parsedName.Suffix != "" {
				parsedName.Suffix += ", " + strings.Join(extraSuffixes, ", ")
			} else {
				parsedName.Suffix = strings.Join(extraSuffixes, ", ")
			}
		}
	}

	// Last name: remove and store last name
	if len(parsedName.nameParts) > 0 {
		parsedName.Last = parsedName.findLastname()
	}

	// First name: remove and store first part as first name
	if len(parsedName.nameParts) > 0 {
		parsedName.First = parsedName.findFirstname()
	}

	// Middle name: store all remaining parts as middle name
	if len(parsedName.nameParts) > 0 {
		parsedName.Middle = parsedName.findMiddlename()
	}

	log.Debugf("Parsing complete: %+v", parsedName)
	return
}

func (parsedName *ParsedName) findNicknames() []string {
	var partsFound []string
	tempString := parsedName.rawName

	matches := nickNameRE.FindAllStringSubmatch(tempString, -1)
	for _, v := range matches {
		partsFound = append(partsFound, v[1])
	}

	log.Debugf("Founded %v nickname(s): %v", len(partsFound), partsFound)
	log.Debug("Clearing")

	for _, v := range matches {
		tempString = strings.Replace(tempString, v[0], "", -1)
	}
	parsedName.rawName = tempString

	log.Debug("Cleared fullname: ", tempString)
	return partsFound
}

func (parsedName *ParsedName) findSuffixes() []string {
	log.Debug("Searching suffixes")
	return parsedName.findParts(suffixList)
}

func (parsedName *ParsedName) findTitles() []string {
	log.Debug("Searching titles")
	return parsedName.findParts(titleList)
}

func (parsedName *ParsedName) splitName() {
	log.Debugf("Spliting fullname: %s", parsedName.rawName)
	fullname := splitNameRE.ReplaceAllLiteralString(parsedName.rawName, " ")
	nameParts := strings.Split(strings.TrimSpace(fullname), " ")
	nameCommas := []bool{}
	for i, v := range nameParts {
		nameParts[i] = strings.TrimSpace(v)
		nameCommas = append(nameCommas, false)
		if strings.HasSuffix(v, ",") {
			nameParts[i] = strings.TrimSuffix(v, ",")
			nameCommas[i] = true
		}
	}

	log.Debug("Splitted parts: ", nameParts)
	log.Debug("Splitted commas: ", nameCommas)
	parsedName.nameParts = nameParts
	parsedName.nameCommas = nameCommas
}

func (parsedName *ParsedName) findParts(list []string) []string {
	var partsFound []string

	for _, namePart := range parsedName.nameParts {
		if namePart == "" {
			continue
		}

		partToCheck := strings.ToLower(namePart)
		partToCheck = strings.TrimSuffix(partToCheck, ".")

		for _, suf := range list {
			if suf == partToCheck {
				partsFound = append(partsFound, namePart)
			}
		}
	}

	log.Debugf("Founded %v parts: %v", len(partsFound), partsFound)
	log.Debug("Clearing")

	for _, partFound := range partsFound {
		foundIndex := -1
		for i, namePart := range parsedName.nameParts {
			if partFound == namePart {
				foundIndex = i
				break
			}
		}
		if foundIndex > -1 {
			parsedName.nameParts = append(parsedName.nameParts[:foundIndex], parsedName.nameParts[foundIndex+1:]...)
			if parsedName.nameCommas[foundIndex] && foundIndex != len(parsedName.nameCommas)-1 {
				parsedName.nameCommas = append(parsedName.nameCommas[:foundIndex+1], parsedName.nameCommas[foundIndex+1+1:]...)
			} else {
				parsedName.nameCommas = append(parsedName.nameCommas[:foundIndex], parsedName.nameCommas[foundIndex+1:]...)
			}
		}
	}

	log.Debug("Cleared parts: ", parsedName.nameParts)
	log.Debug("Cleared commas: ", parsedName.nameCommas)

	return partsFound
}

func (parsedName *ParsedName) joinPrefixes() {
	log.Debug("Join prefixes")

	if len(parsedName.nameParts) > 1 {
		for i := len(parsedName.nameParts) - 2; i >= 0; i-- {
			for _, pref := range prefixList {
				if pref == parsedName.nameParts[i] {
					parsedName.nameParts[i] = parsedName.nameParts[i] + " " + parsedName.nameParts[i+1]
					parsedName.nameParts = append(parsedName.nameParts[:i+1], parsedName.nameParts[i+2:]...)
					parsedName.nameCommas = append(parsedName.nameCommas[:i], parsedName.nameCommas[i+1:]...)
				}
			}
		}
	}

	log.Debug("Prefixes joined: ", strings.Join(parsedName.nameParts, ","))
	log.Debug("Cleared commas: ", parsedName.nameCommas)
}

func (parsedName *ParsedName) joinConjunctions() {
	log.Debug("Join conjunctions")

	if len(parsedName.nameParts) > 2 {
		for i := len(parsedName.nameParts) - 3; i >= 0; i-- {
			for _, conj := range conjunctionList {
				if conj == parsedName.nameParts[i+1] {
					parsedName.nameParts[i] = parsedName.nameParts[i] + " " + parsedName.nameParts[i+1] + " " + parsedName.nameParts[i+2]
					parsedName.nameParts = append(parsedName.nameParts[:i+1], parsedName.nameParts[i+3:]...)
					parsedName.nameCommas = append(parsedName.nameCommas[:i], parsedName.nameCommas[i+2:]...)
					i--
				}
			}
		}
	}
	log.Debug("Conjunctions joined: ", strings.Join(parsedName.nameParts, ","))
	log.Debug("Cleared commas: ", parsedName.nameCommas)
}

func (parsedName *ParsedName) findExtraSuffixes() (extraSuffixes []string) {
	commasCount := 0
	for _, v := range parsedName.nameCommas {
		if v {
			commasCount++
		}
	}
	if commasCount > 1 {
		for i := len(parsedName.nameParts) - 1; i >= 2; i-- {
			if parsedName.nameCommas[i] {
				extraSuffixes = append(extraSuffixes, parsedName.nameParts[i])
				parsedName.nameParts = append(parsedName.nameParts[:i], parsedName.nameParts[i+1:]...)
				parsedName.nameCommas = append(parsedName.nameCommas[:i], parsedName.nameCommas[i+1:]...)
			}
		}
	}

	log.Debugf("Founded %v extra suffixes: %v", len(extraSuffixes), extraSuffixes)
	log.Debug("Cleared commas: ", parsedName.nameCommas)

	return
}

func (parsedName *ParsedName) findLastname() (lastname string) {
	log.Debug("Searching lastname")

	commaIndex := -1
	for i, v := range parsedName.nameCommas {
		if v {
			commaIndex = i
		}
	}

	if commaIndex == -1 {
		commaIndex = len(parsedName.nameParts) - 1
	}

	lastname = parsedName.nameParts[commaIndex]
	parsedName.nameParts = append(parsedName.nameParts[:commaIndex], parsedName.nameParts[commaIndex+1:]...)
	parsedName.nameCommas = parsedName.nameCommas[:0]

	log.Debug("Founded lastname: ", lastname)
	log.Debug("Cleared parts: ", parsedName.nameParts)
	log.Debug("Cleared commas: ", parsedName.nameCommas)
	return
}

func (parsedName *ParsedName) findFirstname() (firstname string) {
	log.Debug("Searching firstname")
	firstname = parsedName.nameParts[0]
	parsedName.nameParts = parsedName.nameParts[1:]
	log.Debug("Founded firstname: ", firstname)
	log.Debug("Cleared parts: ", parsedName.nameParts)
	return
}

func (parsedName *ParsedName) findMiddlename() (middlename string) {
	log.Debug("Searching middlename")
	middlename = strings.Join(parsedName.nameParts, " ")
	parsedName.nameParts = parsedName.nameParts[:0]
	log.Debug("Founded middlename(s): ", middlename)
	log.Debug("Cleared parts: ", parsedName.nameParts)
	return
}
