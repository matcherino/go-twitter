package twitter

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/dghubble/sling"
)

// FriendshipService provides methods for accessing Twitter friendship API
// endpoints.
type FriendshipService struct {
	sling *sling.Sling
}

// newFriendshipService returns a new FriendshipService.
func newFriendshipService(sling *sling.Sling) *FriendshipService {
	return &FriendshipService{
		sling: sling.Path("friendships/"),
	}
}

// FriendshipCreateParams are parameters for FriendshipService.Create
type FriendshipCreateParams struct {
	ScreenName string `url:"screen_name,omitempty"`
	UserID     int64  `url:"user_id,omitempty"`
	Follow     *bool  `url:"follow,omitempty"`
}

// Create creates a friendship to (i.e. follows) the specified user and
// returns the followed user.
// Requires a user auth context.
// https://dev.twitter.com/rest/reference/post/friendships/create
func (s *FriendshipService) Create(params *FriendshipCreateParams) (*User, *http.Response, error) {
	user := new(User)
	apiError := new(APIError)
	resp, err := s.sling.New().Post("create.json").QueryStruct(params).Receive(user, apiError)
	return user, resp, relevantError(err, *apiError)
}

// FriendshipShowParams are paramenters for FriendshipService.Show
type FriendshipShowParams struct {
	SourceID         int64  `url:"source_id,omitempty"`
	SourceScreenName string `url:"source_screen_name,omitempty"`
	TargetID         int64  `url:"target_id,omitempty"`
	TargetScreenName string `url:"target_screen_name,omitempty"`
}

// Show returns the relationship between two arbitrary users.
// Requires a user auth or an app context.
// https://dev.twitter.com/rest/reference/get/friendships/show
func (s *FriendshipService) Show(params *FriendshipShowParams) (*Relationship, *http.Response, error) {
	response := new(RelationshipResponse)
	apiError := new(APIError)
	resp, err := s.sling.New().Get("show.json").QueryStruct(params).Receive(response, apiError)
	return response.Relationship, resp, relevantError(err, *apiError)
}

// RelationshipResponse contains a relationship.
type RelationshipResponse struct {
	Relationship *Relationship `json:"relationship"`
}

// Relationship represents the relation between a source user and target user.
type Relationship struct {
	Source RelationshipSource `json:"source"`
	Target RelationshipTarget `json:"target"`
}

// RelationshipSource represents the source user.
type RelationshipSource struct {
	ID                   int64  `json:"id"`
	IDStr                string `json:"id_str"`
	ScreenName           string `json:"screen_name"`
	Following            bool   `json:"following"`
	FollowedBy           bool   `json:"followed_by"`
	CanDM                bool   `json:"can_dm"`
	Blocking             bool   `json:"blocking"`
	Muting               bool   `json:"muting"`
	AllReplies           bool   `json:"all_replies"`
	WantRetweets         bool   `json:"want_retweets"`
	MarkedSpam           bool   `json:"marked_spam"`
	NotificationsEnabled bool   `json:"notifications_enabled"`
}

// RelationshipTarget represents the target user.
type RelationshipTarget struct {
	ID         int64  `json:"id"`
	IDStr      string `json:"id_str"`
	ScreenName string `json:"screen_name"`
	Following  bool   `json:"following"`
	FollowedBy bool   `json:"followed_by"`
}

// FriendshipDestroyParams are paramenters for FriendshipService.Destroy
type FriendshipDestroyParams struct {
	ScreenName string `url:"screen_name,omitempty"`
	UserID     int64  `url:"user_id,omitempty"`
}

// FriendRelationship represents the follow relationship between a logged in user and a specific users
type FriendRelationship struct {
	Name        string   `json:"name"`
	ScreenName  string   `json:"screen_name"`
	ID          int64    `json:"id"`
	IDStr       string   `json:"id_str"`
	Connections []string `json:"connections"`
}

// Destroy destroys a friendship to (i.e. unfollows) the specified user and
// returns the unfollowed user.
// Requires a user auth context.
// https://dev.twitter.com/rest/reference/post/friendships/destroy
func (s *FriendshipService) Destroy(params *FriendshipDestroyParams) (*User, *http.Response, error) {
	user := new(User)
	apiError := new(APIError)
	resp, err := s.sling.New().Post("destroy.json").QueryStruct(params).Receive(user, apiError)
	return user, resp, relevantError(err, *apiError)
}

// FriendshipPendingParams are paramenters for FriendshipService.Outgoing
type FriendshipPendingParams struct {
	Cursor int64 `url:"cursor,omitempty"`
}

// Outgoing returns a collection of numeric IDs for every protected user for whom the authenticating
// user has a pending follow request.
// https://dev.twitter.com/rest/reference/get/friendships/outgoing
func (s *FriendshipService) Outgoing(params *FriendshipPendingParams) (*FriendIDs, *http.Response, error) {
	ids := new(FriendIDs)
	apiError := new(APIError)
	resp, err := s.sling.New().Get("outgoing.json").QueryStruct(params).Receive(ids, apiError)
	return ids, resp, relevantError(err, *apiError)
}

// Incoming returns a collection of numeric IDs for every user who has a pending request to
// follow the authenticating user.
// https://dev.twitter.com/rest/reference/get/friendships/incoming
func (s *FriendshipService) Incoming(params *FriendshipPendingParams) (*FriendIDs, *http.Response, error) {
	ids := new(FriendIDs)
	apiError := new(APIError)
	resp, err := s.sling.New().Get("incoming.json").QueryStruct(params).Receive(ids, apiError)
	return ids, resp, relevantError(err, *apiError)
}

// FriendshipLookupParams are the parameters for FriendshipService.Lookup
type FriendshipLookupParams struct {
	UserIDList     []int64
	UserIDStr      string `url:"user_id,omitempty"`
	UserID         int64
	ScreenNameList []string
	ScreenNameStr  string `url:"screen_name,omitempty"`
}

// Lookup returns a set of friendship status information between the specified user and a list of users.
// https://developer.twitter.com/en/docs/accounts-and-users/follow-search-get-users/api-reference/get-friendships-lookup
func (s *FriendshipService) Lookup(params *FriendshipLookupParams) ([]FriendRelationship, *http.Response, error) {
	relationships := new([]FriendRelationship)
	transformedParams := new(FriendshipLookupParams)
	apiError := new(APIError)

	// Transform params into a comma separated pair of strings
	transformedParams.ScreenNameStr = strings.Join(params.ScreenNameList, ",")
	if len(transformedParams.ScreenNameStr) > 0 {
		transformedParams.ScreenNameStr += ","
	}
	transformedParams.ScreenNameStr += params.ScreenNameStr

	transformedParams.UserIDStr = params.UserIDStr
	if params.UserID > 0 {
		if len(transformedParams.UserIDStr) > 0 {
			transformedParams.UserIDStr += ","
		}
		transformedParams.UserIDStr += strconv.FormatInt(params.UserID, 10)
	}
	for _, id := range params.UserIDList {
		if len(transformedParams.UserIDStr) > 0 {
			transformedParams.UserIDStr += ","
		}
		transformedParams.UserIDStr += strconv.FormatInt(id, 10)
	}
	users := strings.Count(transformedParams.UserIDStr, ",")
	names := strings.Count(transformedParams.ScreenNameStr, ",")
	if users > 0 {
		users++
	}
	if names > 0 {
		names++
	}
	if names+users > 100 {
		return nil, nil, APIError{
			Errors: []ErrorDetail{
				ErrorDetail{Message: "This API only supports up to 100 users", Code: 200},
			},
		}
	}

	resp, err := s.sling.New().Get("lookup.json").QueryStruct(transformedParams).Receive(relationships, apiError)
	return *relationships, resp, relevantError(err, *apiError)
}
