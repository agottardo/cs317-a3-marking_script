package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

var smtpExecutable *exec.Cmd

const AssignmentDirectory = "/Users/agott/Downloads/a3_jonatan"

func main() {

	log.Println("317 A3 marking script starting.")

	// make clean
	log.Println("running make clean on the student submission first...")
	makeClean := exec.Command("make", "clean")
	makeClean.Dir = AssignmentDirectory
	cleanOut, err := makeClean.Output()
	if err != nil {
		log.Fatalln("Error running make clean on assignment: ", err)
	}
	log.Println("make clean output:")
	log.Print(string(cleanOut))

	// make
	log.Println("okay, now running make...")
	makeCmd := exec.Command("make")
	makeCmd.Dir = AssignmentDirectory
	out, err := makeCmd.Output()
	if err != nil {
		log.Fatalln("Error running Make on assignment: ", err)
	}
	log.Println("Make output:")
	log.Print(string(out))

	// Start the SMTP server in the background, wait for things to go.
	log.Println("Init completed. Starting mysmtpd.")
	go startSMTPServer()
	log.Println("Server started, now sleeping for 3 seconds.")
	time.Sleep(3 * time.Second)

	// Aight. Let's start by opening a simple TCP connection.
	// All submissions should handle this since it was provided in the starter,
	// and allows us to filter out non-functioning implementations quickly.
	localAddr, err := net.ResolveTCPAddr("tcp", ":1001")
	serverAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:50000")
	log.Println("TEST #0: Attempting to connect to mysmtpd. This should work since it was provided in the starter.")
	conn, err := net.DialTCP("tcp", localAddr, serverAddr)
	if err != nil {
		log.Fatalln("⛔ Catastrophic failure. Unable to connect to mysmtpd on port 50000.")
	}

	// Make sure a banner was sent, starting with 220.
	log.Println("TEST #1: Make sure a banner was sent, starting with 220.")
	message, err := bufio.NewReader(conn).ReadString('\r')
	if err != nil {
		log.Fatalln("⛔ Unable to read welcome banner from server:", err)
	}
	if len(message) > 4 && strings.Split(message, " ")[0] == "220" {
		log.Println("👍 It replied with 220 and a welcome banner:", message)
	}

	// Ensure that sending a wrong command returns an error
	log.Println("TEST #2: Sending a wrong command (AAAZ) to check if error codes are handled.")
	io.WriteString(conn, "AAAZ\r\n")
	errMsg, err := bufio.NewReader(conn).ReadString('\r')
	if err != nil {
		log.Fatalln("⛔ Unable to read error from server:", err)
	}
	if len(message) > 2 && strings.HasPrefix(errMsg, "5") {
		log.Println("👍 It replied with error code to a wrong command.")
	} else {
		log.Println("⛔ It didn't reply properly to AAAZ:", message)
	}

	// Ensure that sending a NOOP returns a 2** code
	log.Println("TEST #3: Sending a NOOP to ensure it returns a 2** code.")
	io.WriteString(conn, "NOOP\r\n")
	noopReply, err := bufio.NewReader(conn).ReadString('\r')
	if err != nil {
		log.Fatalln("⛔ Unable to read noop reply from server:", err)
	}
	if len(noopReply) > 2 && strings.HasPrefix(noopReply, "2") {
		log.Println("👍 It replied with 2** code to NOOP.")
	} else {
		log.Println("⛔ It didn't reply properly to NOOP:", noopReply)
	}

	// Ensure that sending a HELO returns a 2** code
	log.Println("TEST #4: Sending 'HELO smtp.gottardo.me' to ensure it returns 2** and the client hostname.")
	io.WriteString(conn, "HELO smtp.gottardo.me\r\n")
	heloReply, err := bufio.NewReader(conn).ReadString('\r')
	if err != nil {
		log.Fatalln("⛔ Unable to read helo reply from server:", err)
	}
	if len(heloReply) > 2 && strings.HasPrefix(heloReply, "2") {
		log.Println("👍 It replied with 2** code to HELO.")
		if strings.Contains(heloReply, "smtp.gottardo.me") {
			log.Println("👍 It replied with the client hostname to HELO.")
		} else {
			log.Println("⛔ It didn't reply with the client hostname to HELO:", heloReply)
		}
	} else {
		log.Println("⛔ It didn't reply properly to HELO:", heloReply)
	}

	// Ensure that it rejects RCPT without sending MAIL before it.
	log.Println("TEST #5: Sending RCPT to ensure it is rejected without MAIL before it.")
	io.WriteString(conn, "RCPT TO:<god@heaven.paradise>\r\n")
	rcptReply, err := bufio.NewReader(conn).ReadString('\r')
	if err != nil {
		log.Fatalln("⛔ Unable to read RCPT reply from server:", err)
	}
	// make sure they don't reply the user doesn't exist (550), that's wrong at this point
	if len(rcptReply) > 2 && strings.HasPrefix(rcptReply, "5") && !strings.HasPrefix(rcptReply, "550") {
		log.Println("👍 It replied with 5** code to RCPT without MAIL before it.")
	} else {
		log.Println("⛔️ It didn't reply properly to RCPT without MAIL before it:", rcptReply)
	}

	// Ensure that it rejects DATA just after HELO.
	log.Println("TEST #6: Sending DATA just after HELO to ensure it is rejected.")
	io.WriteString(conn, "DATA\r\n")
	dataReply, err := bufio.NewReader(conn).ReadString('\r')
	if err != nil {
		log.Fatalln("⛔ Unable to read DATA reply from server:", err)
	}
	if len(dataReply) > 2 && strings.HasPrefix(dataReply, "5") {
		log.Println("👍 It replied with 5** code to DATA right after HELO.")
	} else {
		log.Println("⛔️ It didn't reply properly to DATA after HELO:", dataReply)
	}

	// Ensure that it rejects MAIL with a wrong parameter (AAA) instead of FROM.
	log.Println("TEST #7: Sending 'MAIL AAA' (wrong parameter) to ensure it is rejected.")
	io.WriteString(conn, "MAIL AAA\r\n")
	mail1Reply, err := bufio.NewReader(conn).ReadString('\r')
	if err != nil {
		log.Fatalln("⛔ Unable to read MAIL AAA reply from server:", err)
	}
	if len(mail1Reply) > 2 && strings.HasPrefix(mail1Reply, "5") {
		log.Println("👍 It replied with 5** code to MAIL AAA.")
	} else {
		log.Println("⛔️ It didn't reply properly to MAIL AAA:", mail1Reply)
	}

	conn.Close()

	log.Println("👋 End of marking script. Killing mysmtpd. Goodbye.")
	smtpExecutable.Process.Kill()
}

func startSMTPServer() {
	smtpExecutable = exec.Command("./mysmtpd", "50000")
	smtpExecutable.Dir = AssignmentDirectory
	err := smtpExecutable.Run()
	if err != nil {
		log.Fatalln("⛔ Error returned when running assignment: ", err)
	}
}
