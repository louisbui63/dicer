package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"os/signal"
	"syscall"
	"os"
	"strings"
	"strconv"
	"time"
	"math/rand"
	"unicode"
	"errors"
	"sort"
	"io/ioutil"
)

var precedence map[string]int
func init_parser() {
	precedence = make(map[string]int)
	precedence["r"] = 6
	precedence["d"] = 5
	precedence["x"] = 0
	precedence["k"] = 2
	precedence["K"] = 2
	precedence["s"] = 1
}

func main() {
	btoken, err := ioutil.ReadFile("token.txt")
	if err != nil {
		fmt.Println("fatal error : unable to retrieve token")
		return
	}
	token := string(btoken)

	init_parser()

	rand.Seed(time.Now().UnixNano())

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("fatal error : session couldn't be created")
		return
	}
	discord.AddHandler(handle)
	discord.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages


	err = discord.Open() 
	if err != nil {
		fmt.Println("fatal error : couldn't open connection")
		return
	}
	fmt.Println("bot is now running");

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}

type Token struct {
	ttype string
	value string

}


func isOperator(r rune) bool {
	if r == 'd' || r == 'x' || r == 'k' || r == 'K' || r == 's' {
		return true
	}
	return false
}

func dop(t1 Token, t2 Token) (Token, error) {
	nb, err := strconv.Atoi(t1.value)
	if err != nil {
		return Token{"",""}, errors.New("invalid type provided")
	}
	max, err := strconv.Atoi(t2.value)
	if err != nil {
		return Token{"",""}, errors.New("invalid type provided")
	}

	result := []string{}
	for i := 0; i<nb; i++ {
		result = append(result, strconv.Itoa(rand.Intn(max) +1))
	}
	return Token {"array", strings.Trim(strings.Replace(fmt.Sprint(result), " ", ",", -1), "[]")}, nil
}

func xop(t1 Token, t2 []Token) (Token, error) {
	nb, err := strconv.Atoi(t1.value)
	if err != nil {
		return Token{"",""}, errors.New("invalid type provided")
	}
	ret := ""
	for i := 0; i < nb; i++ {
		ret += t2[i].value
		ret += ";"
	}
	ret = ret[:len(ret)-1]
	return Token{"biarray", ret}, nil
}

func rop(t1 Token, t2 Token) (Token, error) {
	start, err := strconv.Atoi(t1.value)
	if err != nil {
		return Token{"",""}, errors.New("invalid type provided")
	}
	end, err := strconv.Atoi(t2.value)
	if err != nil {
		return Token{"",""}, errors.New("invalid type provided")
	}

	return Token{"number", strconv.Itoa(rand.Intn(end-start+1)+start)}, nil
}

func kop(t1 Token, t2 Token) (Token, error) {
	vals := strings.Split(t1.value, ",")
	qt, err := strconv.Atoi(t2.value)
	if err != nil {
		return Token{"",""}, errors.New("invalid type provided")
	}
	ivals := []int{}
	for i:=0;i<len(vals);i++ {
		v, err := strconv.Atoi(vals[i])
		if err != nil {
			return Token{"",""}, errors.New("invalid type provided")
		}
		ivals = append(ivals, v)
	}
	sort.Ints(ivals)
	result := ivals[:qt]
	return Token {"array", strings.Trim(strings.Replace(fmt.Sprint(result), " ", ",", -1), "[]")}, nil
}

func Kop(t1 Token, t2 Token) (Token, error) {
	vals := strings.Split(t1.value, ",")
	qt, err := strconv.Atoi(t2.value)
	if err != nil {
		return Token{"",""}, errors.New("invalid type provided")
	}
	ivals := []int{}
	for i:=0;i<len(vals);i++ {
		v, err := strconv.Atoi(vals[i])
		if err != nil {
			return Token{"",""}, errors.New("invalid type provided")
		}
		ivals = append(ivals, v)
	}
	sort.Ints(ivals)
	result := ivals[len(ivals)-qt:]
	return Token {"array", strings.Trim(strings.Replace(fmt.Sprint(result), " ", ",", -1), "[]")}, nil
}

func eval(pos int, stack []Token) (Token, int, error) {
	if stack[pos].ttype == "array" || stack[pos].ttype == "number" || stack[pos].ttype == "biarray" {
		return stack[pos], 0, nil
	} else { // op

		switch stack[pos].value {
		case "x":
			op2 := []Token{}
			_, i, err1 := eval(pos-1, stack)
			op1, j, err2 := eval(pos-i-2, stack)
			qt, err := strconv.Atoi(op1.value)

			for i:=0;i<qt;i++ {
				op, _, err := eval(pos-1, stack)
				if err != nil {
					return Token{"",""}, 0, errors.New("err")
				}
				op2 = append(op2, op)
			}
			a, err3 := xop(op1, op2)
			if err1 != nil || err2 != nil || err3 != nil || err != nil {
				return Token{"",""}, 0, errors.New("err")
			}

			return a, 2+i+j, nil
		case "d":
			op2, i, err1 := eval(pos-1, stack)
			op1, j, err2 := eval(pos-i-2, stack)
			a, err3 := dop(op1, op2)
			if err1 != nil || err2 != nil || err3 != nil {
				return Token{"",""}, 0, errors.New("err")
			}
			return a, 2+i+j, nil
		case "r":
			op2, i, err1 := eval(pos-1, stack)
			op1, j, err2 := eval(pos-i-2, stack)
			a, err3 := rop(op1, op2)
			if err1 != nil || err2 != nil || err3 != nil {
				return Token{"",""}, 0, errors.New("err")
			}
			return a, 2+i+j, nil
		case "k":
			op2, i, err1 := eval(pos-1, stack)
			op1, j, err2 := eval(pos-i-2, stack)
			a, err3 := kop(op1, op2)
			if err1 != nil || err2 != nil || err3 != nil {
				return Token{"",""}, 0, errors.New("err")
			}
			return a, 2+i+j, nil
		case "K":
			op2, i, err1 := eval(pos-1, stack)
			op1, j, err2 := eval(pos-i-2, stack)
			a, err3 := Kop(op1, op2)
			if err1 != nil || err2 != nil || err3 != nil {
				return Token{"",""}, 0, errors.New("err")
			}
			return a, 2+i+j, nil
		case "s":
			op, i, err1 := eval(pos-1, stack)
			if err1 != nil {
				return Token{"",""}, 0, errors.New("err")
			}
			el := strings.Split(op.value, ",")
			a := 0
			for i:=0;i<len(el);i++ {
				b, err := strconv.Atoi(el[i])
				if err != nil {
					return Token{"",""}, 0, errors.New("err")
				}
				a += b
			}
			return Token{"number", strconv.Itoa(a)}, 1+i, nil
		}
	}
	return stack[pos], 0, nil
}

func handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return // we don't want to answer ourself
	}
	if m.Content == "!dice clear" {
		if m.Author.Username == "louisbui63" {
			id := m.ChannelID	
			guild := m.GuildID
			channels, err := s.GuildChannels(guild)
			if err != nil {
				fmt.Println("WHAAAAT!!!!")
			}
			for _, c := range channels {
				if c.ID == id {
					ms, err := s.ChannelMessages(id, -1, "", "", "")
					if err != nil {
						continue
					}
					mss := []string{}
					for _, i := range ms {
						mss = append(mss, i.ID)
					}
					s.ChannelMessagesBulkDelete(id, mss)
					break
				}
			}
		}

	} else if m.Content == "!dice help" || m.Content == "!dicer help" {
		s.ChannelMessageSend(m.ChannelID, `Dicer is a dice roller bot designed for tabletop rpg. It is based on an innovative representation of rolls as mathematical expressions, allowing endless possibilities, end thus making it suitable no matter the rules you are using.
	the following operators are available :
		- ndm rolls n m-sized dices and put the results in an array
		- nxm repeats n times m and put everything in a vertical array
		- nrm takes a random integer between n and m, both included
	!dice followed by a command outputs the result of this command.
	there also some specific commands :
	!dice help		: displays this help
	!dice clear		: clears the channel`)

	} else if strings.HasPrefix(m.Content, "!dice ") {
		expr := m.Content[6:]
		// check if not clear
		tokens := []Token{}

		currentType := ""
		buffer := ""
		for i := 0; i < len(expr); i++ {
			if unicode.IsDigit(rune(expr[i])) {
				if currentType != "number" && currentType != "" {
					tokens = append(tokens, Token{currentType, buffer})
					buffer = ""
				}
				currentType = "number"
				buffer += string(expr[i])

			} else if isOperator(rune(expr[i])) {
				if currentType != "" {
					tokens = append(tokens, Token{currentType, buffer})
					buffer = ""
				}
				currentType = "operator"
				buffer += string(expr[i])
			} else if expr[i] == '(' {
				if currentType != "" {
					tokens = append(tokens, Token{currentType, buffer})
					buffer = ""
				}
				currentType = "lparen"
				buffer += string(expr[i])
			} else if expr[i] == ')' {
				if currentType != "" {
					tokens = append(tokens, Token{currentType, buffer})
					buffer = ""
				}
				currentType = "rparen"
				buffer += string(expr[i])
			}

		}
		tokens = append(tokens, Token{currentType, buffer})

		oustack := []Token{}
		opstack := []Token{}
		for i := 0; i < len(tokens); i++ {
			switch tokens[i].ttype {
			case "number":
				oustack = append(oustack, tokens[i])
			case "operator":
				for len(opstack) > 0 && (opstack[len(opstack)-1].ttype == "operator" && (precedence[opstack[len(opstack)-1].value] > precedence[tokens[i].value] || precedence[opstack[len(opstack)-1].value] /*should also be leftasso*/== precedence[tokens[i].value])) {
					oustack = append(oustack, opstack[len(opstack)-1])
					opstack = opstack[:len(opstack)-1]
				}
				opstack = append(opstack, tokens[i])
			case "lparen":
				opstack = append(opstack, tokens[i])
			case "rparen":
				for (opstack[len(opstack) -1].ttype != "lparen") {
					oustack = append(oustack, opstack[len(opstack)-1])
					opstack = opstack[:len(opstack)-1]
				}
				opstack = opstack[:len(opstack)-1]
			}
		}
		for len(opstack) != 0 {
			oustack = append(oustack, opstack[len(opstack)-1])
			opstack = opstack[:len(opstack)-1]
		}

		fmt.Println(oustack)

		out, _, err := eval( len(oustack)-1, oustack[:])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Malformed request")
			return
		}

		fmt.Println(out.value)

		a := strings.Split(out.value, ";")
		t := [][]string{}
		for i := 0; i < len(a); i++ {
			t = append(t, strings.Split(a[i], ","))
		}

		for i:=0; i<len(t);i++{
			/*for j:=0; j<len(t[i]);j++{
				_, err := strconv.Atoi(t[i][j])
				if err != nil {
					s := strings.Split(t[i][j], "r")
					s0, _ := strconv.Atoi(s[0])
					s1, _ := strconv.Atoi(s[1])
					t[i][j] = strconv.Itoa(rand.Intn(s1-s0+1)+s0)
				}
			}*/
			mess := strings.Trim(strings.Replace(fmt.Sprint(t[i]), " ", ",", -1), "[]")
			s.ChannelMessageSend(m.ChannelID, mess)
		}


	}
}


/*func handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return // we don't want to answer ourself
	}
	if m.Content == "!dice 1dSATAN" {
		s.ChannelMessageSend(m.ChannelID, "666")
	} else if m.Content == "!dice 1dCALLTHEPOLICE" {
		s.ChannelMessageSend(m.ChannelID, "911")
	} else if m.Content == "!dice 1dTHEANSWERTOLIFEANDEVERYTHING" {
		s.ChannelMessageSend(m.ChannelID, "42")
	} else if m.Content == "!dice 1dALEXIS" {
		s.ChannelMessageSend(m.ChannelID, "17 (seulement)")
	} else if m.Content == "!dice 1dMAXENCER" {
		s.ChannelMessageSend(m.ChannelID, "1,60m")
	} else if m.Content == "!dice clear" {
		if m.Author.Username == "louisbui63" {
			id := m.ChannelID	
			guild := m.GuildID
			channels, err := s.GuildChannels(guild)
			if err != nil {
				fmt.Println("WHAAAAT!!!!")
			}
			for _, c := range channels {
				if c.ID == id {
					ms, err := s.ChannelMessages(id, -1, "", "", "")
					if err != nil {
						continue
					}
					mss := []string{}
					for _, i := range ms {
						mss = append(mss, i.ID)
					}
					s.ChannelMessagesBulkDelete(id, mss)
					break
				}
			}
		}
	} else if strings.HasPrefix(m.Content, "!dice ") {
		command := m.Content[6:]
		b := strings.Split(command, "x")
		nb := 1
		if len(b) == 2 {
			nb1, err := strconv.Atoi(b[0])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Malformed request")
				return
			}
			nb = nb1
			command = b[1]
		}
		b = strings.Split(command, "]+")
		tadder := 0
		if len(b) == 2 {
			tadder1, err := strconv.Atoi(b[1])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Malformed request")
				return
			}
			tadder = tadder1
			command = b[0]
		}
		b = strings.Split(command, "+")
		adder := 0
		if len(b) == 2 {
			adder1, err := strconv.Atoi(b[1])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Malformed request")
				return
			}
			adder = adder1
			command = b[0]
		}
		for j := 0; j<nb; j++ {
			a := strings.Split(command, "d")
			a0, err := strconv.Atoi(a[0])
			if err != nil || a0 < 0 || a0 > 20 {
				s.ChannelMessageSend(m.ChannelID, "Malformed request")
				return
			}
			a1, err := strconv.Atoi(a[1])
			if err != nil || a1 <= 0 || a1 > 1000 {
				s.ChannelMessageSend(m.ChannelID, "Malformed request")
				return
			}
			list := []int{}
			for i := 0; i < a0; i++ {
				list = append(list, rand.Intn(a1)+1+adder)
			}
			tot := 0
			for i := 0; i< len(list); i++ {
				tot += list[i]
			}
			mess := strings.Trim(strings.Replace(fmt.Sprint(list), " ", ",", -1), "[]")
			mess += (" => " + fmt.Sprint(tot))
			if tadder != 0 {
				mess += " => " + fmt.Sprint(tot+tadder)
			}
			s.ChannelMessageSend(m.ChannelID, mess)
		}
	}
}*/
