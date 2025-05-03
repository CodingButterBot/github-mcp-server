package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListProjects lists projects for an organization, repository, or user.
// Since Projects V2 uses GraphQL, we have to make direct GraphQL queries.
func ListProjects(getClient GetClientFn, t translations.TranslationHelperFunc) (mcp.Tool, mcp.ToolHandlerFunc) {
	tool := mcp.NewTool(
		"list_projects",
		t("TOOL_LIST_PROJECTS_DESCRIPTION", "List projects (v2) for an organization, user, or repository"),
		mcp.WithString("owner", mcp.Description("Organization or user login name"), mcp.Required()),
		mcp.WithString("type", mcp.Description("Type of owner (organization, user)"), mcp.Required(), mcp.Enum("organization", "user")),
		WithPagination(),
	)

	handler := func(ctx context.Context, r mcp.CallToolRequest) (mcp.ToolResult, error) {
		client, err := getClient(ctx)
		if err != nil {
			return nil, err
		}

		owner, err := requiredParam[string](r, "owner")
		if err != nil {
			return nil, err
		}

		ownerType, err := requiredParam[string](r, "type")
		if err != nil {
			return nil, err
		}

		pagination, err := OptionalPaginationParams(r)
		if err != nil {
			return nil, err
		}

		// GraphQL query is different depending on owner type
		var query string
		if ownerType == "organization" {
			query = fmt.Sprintf(`
			{
				organization(login: "%s") {
					projectsV2(first: %d) {
						nodes {
							id
							number
							title
							shortDescription
							public
							closed
							url
							createdAt
							updatedAt
						}
						pageInfo {
							hasNextPage
							endCursor
						}
						totalCount
					}
				}
			}
			`, owner, pagination.perPage)
		} else if ownerType == "user" {
			query = fmt.Sprintf(`
			{
				user(login: "%s") {
					projectsV2(first: %d) {
						nodes {
							id
							number
							title
							shortDescription
							public
							closed
							url
							createdAt
							updatedAt
						}
						pageInfo {
							hasNextPage
							endCursor
						}
						totalCount
					}
				}
			}
			`, owner, pagination.perPage)
		} else {
			return nil, fmt.Errorf("invalid owner type: %s", ownerType)
		}

		// Execute GraphQL query
		req, err := client.NewRequest("POST", "graphql", map[string]interface{}{
			"query": query,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating GraphQL request: %w", err)
		}

		var response map[string]interface{}
		_, err = client.Do(ctx, req, &response)
		if err != nil {
			return nil, fmt.Errorf("error executing GraphQL query: %w", err)
		}

		// Convert response to JSON string
		jsonData, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("error marshaling response to JSON: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}

	return tool, handler
}

// GetProject gets details of a specific project.
func GetProject(getClient GetClientFn, t translations.TranslationHelperFunc) (mcp.Tool, mcp.ToolHandlerFunc) {
	tool := mcp.NewTool(
		"get_project",
		t("TOOL_GET_PROJECT_DESCRIPTION", "Get details of a GitHub project (v2)"),
		mcp.WithString("owner", mcp.Description("Organization or user login name"), mcp.Required()),
		mcp.WithString("type", mcp.Description("Type of owner (organization, user)"), mcp.Required(), mcp.Enum("organization", "user")),
		mcp.WithNumber("number", mcp.Description("The project number"), mcp.Required()),
	)

	handler := func(ctx context.Context, r mcp.CallToolRequest) (mcp.ToolResult, error) {
		client, err := getClient(ctx)
		if err != nil {
			return nil, err
		}

		owner, err := requiredParam[string](r, "owner")
		if err != nil {
			return nil, err
		}

		ownerType, err := requiredParam[string](r, "type")
		if err != nil {
			return nil, err
		}

		number, err := RequiredInt(r, "number")
		if err != nil {
			return nil, err
		}

		// GraphQL query is different depending on owner type
		var query string
		if ownerType == "organization" {
			query = fmt.Sprintf(`
			{
				organization(login: "%s") {
					projectV2(number: %d) {
						id
						number
						title
						shortDescription
						public
						closed
						url
						creator {
							login
						}
						createdAt
						updatedAt
						items(first: 20) {
							nodes {
								id
								content {
									... on Issue {
										title
										number
										repository {
											name
										}
									}
									... on PullRequest {
										title
										number
										repository {
											name
										}
									}
								}
							}
						}
						fields(first: 20) {
							nodes {
								... on ProjectV2Field {
									id
									name
									dataType
								}
								... on ProjectV2IterationField {
									id
									name
									dataType
									configuration {
										iterations {
											startDate
											duration
										}
									}
								}
								... on ProjectV2SingleSelectField {
									id
									name
									dataType
									options {
										id
										name
										color
									}
								}
							}
						}
					}
				}
			}
			`, owner, number)
		} else if ownerType == "user" {
			query = fmt.Sprintf(`
			{
				user(login: "%s") {
					projectV2(number: %d) {
						id
						number
						title
						shortDescription
						public
						closed
						url
						creator {
							login
						}
						createdAt
						updatedAt
						items(first: 20) {
							nodes {
								id
								content {
									... on Issue {
										title
										number
										repository {
											name
										}
									}
									... on PullRequest {
										title
										number
										repository {
											name
										}
									}
								}
							}
						}
						fields(first: 20) {
							nodes {
								... on ProjectV2Field {
									id
									name
									dataType
								}
								... on ProjectV2IterationField {
									id
									name
									dataType
									configuration {
										iterations {
											startDate
											duration
										}
									}
								}
								... on ProjectV2SingleSelectField {
									id
									name
									dataType
									options {
										id
										name
										color
									}
								}
							}
						}
					}
				}
			}
			`, owner, number)
		} else {
			return nil, fmt.Errorf("invalid owner type: %s", ownerType)
		}

		// Execute GraphQL query
		req, err := client.NewRequest("POST", "graphql", map[string]interface{}{
			"query": query,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating GraphQL request: %w", err)
		}

		var response map[string]interface{}
		_, err = client.Do(ctx, req, &response)
		if err != nil {
			return nil, fmt.Errorf("error executing GraphQL query: %w", err)
		}

		// Convert response to JSON string
		jsonData, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("error marshaling response to JSON: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}

	return tool, handler
}

// CreateProject creates a new project.
func CreateProject(getClient GetClientFn, t translations.TranslationHelperFunc) (mcp.Tool, mcp.ToolHandlerFunc) {
	tool := mcp.NewTool(
		"create_project",
		t("TOOL_CREATE_PROJECT_DESCRIPTION", "Create a new GitHub project (v2)"),
		mcp.WithString("owner", mcp.Description("Organization or user login name"), mcp.Required()),
		mcp.WithString("type", mcp.Description("Type of owner (organization, user)"), mcp.Required(), mcp.Enum("organization", "user")),
		mcp.WithString("title", mcp.Description("Project title"), mcp.Required()),
		mcp.WithString("description", mcp.Description("Project description"), mcp.Required()),
	)

	handler := func(ctx context.Context, r mcp.CallToolRequest) (mcp.ToolResult, error) {
		client, err := getClient(ctx)
		if err != nil {
			return nil, err
		}

		owner, err := requiredParam[string](r, "owner")
		if err != nil {
			return nil, err
		}

		ownerType, err := requiredParam[string](r, "type")
		if err != nil {
			return nil, err
		}

		title, err := requiredParam[string](r, "title")
		if err != nil {
			return nil, err
		}

		description, err := requiredParam[string](r, "description")
		if err != nil {
			return nil, err
		}

		// First get the owner ID using GraphQL
		var ownerIdQuery string
		if ownerType == "organization" {
			ownerIdQuery = fmt.Sprintf(`
			{
				organization(login: "%s") {
					id
				}
			}
			`, owner)
		} else if ownerType == "user" {
			ownerIdQuery = fmt.Sprintf(`
			{
				user(login: "%s") {
					id
				}
			}
			`, owner)
		} else {
			return nil, fmt.Errorf("invalid owner type: %s", ownerType)
		}

		// Execute GraphQL query to get owner ID
		req, err := client.NewRequest("POST", "graphql", map[string]interface{}{
			"query": ownerIdQuery,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating GraphQL request: %w", err)
		}

		var ownerIdResponse map[string]interface{}
		_, err = client.Do(ctx, req, &ownerIdResponse)
		if err != nil {
			return nil, fmt.Errorf("error executing GraphQL query: %w", err)
		}

		// Extract owner ID from response
		var ownerId string
		if ownerType == "organization" {
			if data, ok := ownerIdResponse["data"].(map[string]interface{}); ok {
				if org, ok := data["organization"].(map[string]interface{}); ok {
					if id, ok := org["id"].(string); ok {
						ownerId = id
					}
				}
			}
		} else if ownerType == "user" {
			if data, ok := ownerIdResponse["data"].(map[string]interface{}); ok {
				if user, ok := data["user"].(map[string]interface{}); ok {
					if id, ok := user["id"].(string); ok {
						ownerId = id
					}
				}
			}
		}

		if ownerId == "" {
			return nil, fmt.Errorf("failed to get owner ID")
		}

		// Create project using GraphQL mutation
		createMutation := fmt.Sprintf(`
		mutation {
			createProjectV2(input: {
				ownerId: "%s",
				title: "%s",
				repositoryId: null,
				description: "%s"
			}) {
				projectV2 {
					id
					number
					title
					url
				}
			}
		}
		`, ownerId, escapeString(title), escapeString(description))

		// Execute GraphQL mutation
		req, err = client.NewRequest("POST", "graphql", map[string]interface{}{
			"query": createMutation,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating GraphQL request: %w", err)
		}

		var createResponse map[string]interface{}
		_, err = client.Do(ctx, req, &createResponse)
		if err != nil {
			return nil, fmt.Errorf("error executing GraphQL mutation: %w", err)
		}

		// Convert response to JSON string
		jsonData, err := json.Marshal(createResponse)
		if err != nil {
			return nil, fmt.Errorf("error marshaling response to JSON: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}

	return tool, handler
}

// AddProjectItem adds an item (issue or PR) to a project.
func AddProjectItem(getClient GetClientFn, t translations.TranslationHelperFunc) (mcp.Tool, mcp.ToolHandlerFunc) {
	tool := mcp.NewTool(
		"add_project_item",
		t("TOOL_ADD_PROJECT_ITEM_DESCRIPTION", "Add an issue or pull request to a GitHub project (v2)"),
		mcp.WithString("project_id", mcp.Description("The project node ID (not number)"), mcp.Required()),
		mcp.WithString("content_id", mcp.Description("The node ID of the issue or pull request"), mcp.Required()),
	)

	handler := func(ctx context.Context, r mcp.CallToolRequest) (mcp.ToolResult, error) {
		client, err := getClient(ctx)
		if err != nil {
			return nil, err
		}

		projectId, err := requiredParam[string](r, "project_id")
		if err != nil {
			return nil, err
		}

		contentId, err := requiredParam[string](r, "content_id")
		if err != nil {
			return nil, err
		}

		// Add item to project using GraphQL mutation
		mutation := fmt.Sprintf(`
		mutation {
			addProjectV2ItemById(input: {
				projectId: "%s",
				contentId: "%s"
			}) {
				item {
					id
				}
			}
		}
		`, projectId, contentId)

		// Execute GraphQL mutation
		req, err := client.NewRequest("POST", "graphql", map[string]interface{}{
			"query": mutation,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating GraphQL request: %w", err)
		}

		var response map[string]interface{}
		_, err = client.Do(ctx, req, &response)
		if err != nil {
			return nil, fmt.Errorf("error executing GraphQL mutation: %w", err)
		}

		// Convert response to JSON string
		jsonData, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("error marshaling response to JSON: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}

	return tool, handler
}

// GetProjectItems gets items in a project.
func GetProjectItems(getClient GetClientFn, t translations.TranslationHelperFunc) (mcp.Tool, mcp.ToolHandlerFunc) {
	tool := mcp.NewTool(
		"get_project_items",
		t("TOOL_GET_PROJECT_ITEMS_DESCRIPTION", "Get items in a GitHub project (v2)"),
		mcp.WithString("project_id", mcp.Description("The project node ID (not number)"), mcp.Required()),
		WithPagination(),
	)

	handler := func(ctx context.Context, r mcp.CallToolRequest) (mcp.ToolResult, error) {
		client, err := getClient(ctx)
		if err != nil {
			return nil, err
		}

		projectId, err := requiredParam[string](r, "project_id")
		if err != nil {
			return nil, err
		}

		pagination, err := OptionalPaginationParams(r)
		if err != nil {
			return nil, err
		}

		// Get project items using GraphQL query
		query := fmt.Sprintf(`
		{
			node(id: "%s") {
				... on ProjectV2 {
					items(first: %d) {
						nodes {
							id
							content {
								... on Issue {
									title
									number
									repository {
										name
									}
								}
								... on PullRequest {
									title
									number
									repository {
										name
									}
								}
							}
							fieldValues(first: 20) {
								nodes {
									... on ProjectV2ItemFieldTextValue {
										text
										field {
											... on ProjectV2FieldCommon {
												name
											}
										}
									}
									... on ProjectV2ItemFieldDateValue {
										date
										field {
											... on ProjectV2FieldCommon {
												name
											}
										}
									}
									... on ProjectV2ItemFieldSingleSelectValue {
										name
										field {
											... on ProjectV2FieldCommon {
												name
											}
										}
									}
								}
							}
						}
						pageInfo {
							hasNextPage
							endCursor
						}
						totalCount
					}
				}
			}
		}
		`, projectId, pagination.perPage)

		// Execute GraphQL query
		req, err := client.NewRequest("POST", "graphql", map[string]interface{}{
			"query": query,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating GraphQL request: %w", err)
		}

		var response map[string]interface{}
		_, err = client.Do(ctx, req, &response)
		if err != nil {
			return nil, fmt.Errorf("error executing GraphQL query: %w", err)
		}

		// Convert response to JSON string
		jsonData, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("error marshaling response to JSON: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}

	return tool, handler
}

// UpdateProjectField updates a field value for a project item.
func UpdateProjectField(getClient GetClientFn, t translations.TranslationHelperFunc) (mcp.Tool, mcp.ToolHandlerFunc) {
	tool := mcp.NewTool(
		"update_project_field",
		t("TOOL_UPDATE_PROJECT_FIELD_DESCRIPTION", "Update a field value for a project item"),
		mcp.WithString("project_id", mcp.Description("The project node ID (not number)"), mcp.Required()),
		mcp.WithString("item_id", mcp.Description("The project item node ID"), mcp.Required()),
		mcp.WithString("field_id", mcp.Description("The project field node ID"), mcp.Required()),
		mcp.WithString("field_type", mcp.Description("The type of field (text, date, number, single_select)"), mcp.Required(), mcp.Enum("text", "date", "number", "single_select")),
		mcp.WithString("value", mcp.Description("The value to set for the field"), mcp.Required()),
		mcp.WithString("option_id", mcp.Description("The option ID for single_select fields"), mcp.Optional()),
	)

	handler := func(ctx context.Context, r mcp.CallToolRequest) (mcp.ToolResult, error) {
		client, err := getClient(ctx)
		if err != nil {
			return nil, err
		}

		projectId, err := requiredParam[string](r, "project_id")
		if err != nil {
			return nil, err
		}

		itemId, err := requiredParam[string](r, "item_id")
		if err != nil {
			return nil, err
		}

		fieldId, err := requiredParam[string](r, "field_id")
		if err != nil {
			return nil, err
		}

		fieldType, err := requiredParam[string](r, "field_type")
		if err != nil {
			return nil, err
		}

		value, err := requiredParam[string](r, "value")
		if err != nil {
			return nil, err
		}

		// Build mutation based on field type
		var mutation string
		switch fieldType {
		case "text":
			mutation = fmt.Sprintf(`
			mutation {
				updateProjectV2ItemFieldValue(input: {
					projectId: "%s",
					itemId: "%s",
					fieldId: "%s",
					value: { text: "%s" }
				}) {
					projectV2Item {
						id
					}
				}
			}
			`, projectId, itemId, fieldId, escapeString(value))
		case "date":
			mutation = fmt.Sprintf(`
			mutation {
				updateProjectV2ItemFieldValue(input: {
					projectId: "%s",
					itemId: "%s",
					fieldId: "%s",
					value: { date: "%s" }
				}) {
					projectV2Item {
						id
					}
				}
			}
			`, projectId, itemId, fieldId, value)
		case "number":
			mutation = fmt.Sprintf(`
			mutation {
				updateProjectV2ItemFieldValue(input: {
					projectId: "%s",
					itemId: "%s",
					fieldId: "%s",
					value: { number: %s }
				}) {
					projectV2Item {
						id
					}
				}
			}
			`, projectId, itemId, fieldId, value)
		case "single_select":
			optionId, err := requiredParam[string](r, "option_id")
			if err != nil {
				return nil, fmt.Errorf("option_id is required for single_select fields: %w", err)
			}
			mutation = fmt.Sprintf(`
			mutation {
				updateProjectV2ItemFieldValue(input: {
					projectId: "%s",
					itemId: "%s",
					fieldId: "%s",
					value: { singleSelectOptionId: "%s" }
				}) {
					projectV2Item {
						id
					}
				}
			}
			`, projectId, itemId, fieldId, optionId)
		default:
			return nil, fmt.Errorf("unsupported field type: %s", fieldType)
		}

		// Execute GraphQL mutation
		req, err := client.NewRequest("POST", "graphql", map[string]interface{}{
			"query": mutation,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating GraphQL request: %w", err)
		}

		var response map[string]interface{}
		_, err = client.Do(ctx, req, &response)
		if err != nil {
			return nil, fmt.Errorf("error executing GraphQL mutation: %w", err)
		}

		// Convert response to JSON string
		jsonData, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("error marshaling response to JSON: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}

	return tool, handler
}

// Helper function to escape strings for GraphQL
func escapeString(s string) string {
	return strings.Replace(
		strings.Replace(
			strings.Replace(
				strings.Replace(s, "\\", "\\\\", -1),
				"\"", "\\\"", -1),
			"\n", "\\n", -1),
		"\r", "\\r", -1)
}

// GetContentId gets the node ID of an issue or pull request.
func GetContentId(getClient GetClientFn, t translations.TranslationHelperFunc) (mcp.Tool, mcp.ToolHandlerFunc) {
	tool := mcp.NewTool(
		"get_content_id",
		t("TOOL_GET_CONTENT_ID_DESCRIPTION", "Get the node ID of an issue or pull request for use with Projects API"),
		mcp.WithString("owner", mcp.Description("Repository owner"), mcp.Required()),
		mcp.WithString("repo", mcp.Description("Repository name"), mcp.Required()),
		mcp.WithNumber("number", mcp.Description("Issue or pull request number"), mcp.Required()),
		mcp.WithString("type", mcp.Description("Content type (issue or pull_request)"), mcp.Required(), mcp.Enum("issue", "pull_request")),
	)

	handler := func(ctx context.Context, r mcp.CallToolRequest) (mcp.ToolResult, error) {
		client, err := getClient(ctx)
		if err != nil {
			return nil, err
		}

		owner, err := requiredParam[string](r, "owner")
		if err != nil {
			return nil, err
		}

		repo, err := requiredParam[string](r, "repo")
		if err != nil {
			return nil, err
		}

		number, err := RequiredInt(r, "number")
		if err != nil {
			return nil, err
		}

		contentType, err := requiredParam[string](r, "type")
		if err != nil {
			return nil, err
		}

		// GraphQL query to get node ID
		var query string
		if contentType == "issue" {
			query = fmt.Sprintf(`
			{
				repository(owner: "%s", name: "%s") {
					issue(number: %d) {
						id
					}
				}
			}
			`, owner, repo, number)
		} else if contentType == "pull_request" {
			query = fmt.Sprintf(`
			{
				repository(owner: "%s", name: "%s") {
					pullRequest(number: %d) {
						id
					}
				}
			}
			`, owner, repo, number)
		} else {
			return nil, fmt.Errorf("invalid content type: %s", contentType)
		}

		// Execute GraphQL query
		req, err := client.NewRequest("POST", "graphql", map[string]interface{}{
			"query": query,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating GraphQL request: %w", err)
		}

		var response map[string]interface{}
		_, err = client.Do(ctx, req, &response)
		if err != nil {
			return nil, fmt.Errorf("error executing GraphQL query: %w", err)
		}

		// Convert response to JSON string
		jsonData, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("error marshaling response to JSON: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}

	return tool, handler
}