package integration_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/wundergraph/cosmo/router/core"
	"github.com/wundergraph/cosmo/router/pkg/config"
	"net/http"
	"testing"

	"github.com/wundergraph/cosmo/router-tests/testenv"
)

func TestSingleFileUpload(t *testing.T) {
	t.Parallel()
	testenv.Run(t, &testenv.Config{}, func(t *testing.T, xEnv *testenv.Environment) {
		files := make([][]byte, 1)
		files[0] = []byte("File content as text")
		res := xEnv.MakeGraphQLRequestOK(testenv.GraphQLRequest{
			Query:     "mutation ($file: Upload!){singleUpload(file: $file)}",
			Variables: []byte(`{"file":null}`),
			Files:     files,
		})
		require.JSONEq(t, `{"data":{"singleUpload": true}}`, res.Body)
	})
}

func TestSingleFileUpload_InvalidFileFormat(t *testing.T) {
	t.Parallel()
	testenv.Run(t, &testenv.Config{}, func(t *testing.T, xEnv *testenv.Environment) {
		res := xEnv.MakeGraphQLRequestOK(testenv.GraphQLRequest{
			Query:     "mutation ($file: Upload!){singleUpload(file: $file)}",
			Variables: []byte(`{"file":"invalid_format"}`),
		})
		require.JSONEq(t, `{"errors":[{"message":"Failed to fetch from Subgraph '0' at Path 'mutation'.","extensions":{"errors":[{"message":"string is not an Upload","path":["singleUpload","file"]}],"statusCode":200}},{"message":"Cannot return null for non-nullable field 'Mutation.singleUpload'.","path":["singleUpload"]}],"data":null}`, res.Body)
	})
}

func TestSingleFileUpload_NoFileProvided(t *testing.T) {
	t.Parallel()
	testenv.Run(t, &testenv.Config{}, func(t *testing.T, xEnv *testenv.Environment) {
		res := xEnv.MakeGraphQLRequestOK(testenv.GraphQLRequest{
			Query:     "mutation ($file: Upload!){singleUpload(file: $file)}",
			Variables: []byte(`{"file":null}`),
		})
		require.JSONEq(t, `{"errors":[{"message":"Failed to fetch from Subgraph '0' at Path 'mutation'.","extensions":{"errors":[{"message":"cannot be null","path":["variable","file"],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"statusCode":422}},{"message":"Cannot return null for non-nullable field 'Mutation.singleUpload'.","path":["singleUpload"]}],"data":null}`, res.Body)
	})
}

func TestSingleFileUpload_FileSizeExceedsLimit(t *testing.T) {
	t.Parallel()
	testenv.Run(t, &testenv.Config{
		RouterOptions: []core.Option{core.WithRouterTrafficConfig(&config.RouterTrafficConfiguration{
			MaxUploadRequestBodyBytes: 100,
		})},
	}, func(t *testing.T, xEnv *testenv.Environment) {
		files := make([][]byte, 1)
		files[0] = []byte("This is an example of a large file that exceeds the max request body size.")
		res, err := xEnv.MakeGraphQLRequest(testenv.GraphQLRequest{
			Query:     "mutation ($file: Upload!){singleUpload(file: $file)}",
			Variables: []byte(`{"file":null}`),
			Files:     files,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusRequestEntityTooLarge, res.Response.StatusCode)
		require.Equal(t, `{"errors":[{"message":"request body too large"}],"data":null}`, res.Body)
	})
}

func TestMultipleFilesUpload(t *testing.T) {
	t.Parallel()
	testenv.Run(t, &testenv.Config{}, func(t *testing.T, xEnv *testenv.Environment) {
		files := make([][]byte, 2)
		files[0] = []byte("File1 content as text")
		files[1] = []byte("File2 content as text")
		res := xEnv.MakeGraphQLRequestOK(testenv.GraphQLRequest{
			Query:     "mutation($files: [Upload!]!) { multipleUpload(files: $files)}",
			Variables: []byte(`{"files":[null, null]}`),
			Files:     files,
		})
		require.JSONEq(t, `{"data":{"multipleUpload": true}}`, res.Body)
	})
}

func TestMultipleFilesUpload_InvalidFileFormat(t *testing.T) {
	t.Parallel()
	testenv.Run(t, &testenv.Config{}, func(t *testing.T, xEnv *testenv.Environment) {
		res := xEnv.MakeGraphQLRequestOK(testenv.GraphQLRequest{
			Query:     "mutation($files: [Upload!]!) { multipleUpload(files: $files)}",
			Variables: []byte(`{"files":["invalid_format1", "invalid_format2"]}`),
		})
		require.JSONEq(t, `{"errors":[{"message":"Failed to fetch from Subgraph '0' at Path 'mutation'.","extensions":{"errors":[{"message":"string is not an Upload","path":["multipleUpload","files",0]}],"statusCode":200}},{"message":"Cannot return null for non-nullable field 'Mutation.multipleUpload'.","path":["multipleUpload"]}],"data":null}`, res.Body)
	})
}

func TestMultipleFilesUpload_NoFilesProvided(t *testing.T) {
	t.Parallel()
	testenv.Run(t, &testenv.Config{}, func(t *testing.T, xEnv *testenv.Environment) {
		res := xEnv.MakeGraphQLRequestOK(testenv.GraphQLRequest{
			Query:     "mutation($files: [Upload!]!) { multipleUpload(files: $files)}",
			Variables: []byte(`{"files":null}`),
		})
		fmt.Println(res.Body)
		require.JSONEq(t, `{"errors":[{"message":"Failed to fetch from Subgraph '0' at Path 'mutation'.","extensions":{"errors":[{"message":"could not render fetch input","path":[]}]}},{"message":"Cannot return null for non-nullable field 'Mutation.multipleUpload'.","path":["multipleUpload"]}],"data":null}`, res.Body)
	})
}

func TestMultipleFilesUpload_FileSizeExceedsLimit(t *testing.T) {
	t.Parallel()
	testenv.Run(t, &testenv.Config{
		RouterOptions: []core.Option{core.WithRouterTrafficConfig(&config.RouterTrafficConfiguration{
			MaxUploadRequestBodyBytes: 100,
		})},
	}, func(t *testing.T, xEnv *testenv.Environment) {
		files := make([][]byte, 2)
		files[0] = []byte("This is an example of a large file that exceeds the max request body size.")
		files[1] = []byte("Another file.")
		res, err := xEnv.MakeGraphQLRequest(testenv.GraphQLRequest{
			Header: map[string][]string{
				"Content-Type": {"multipart/form-data"},
			},
			Query:     "mutation($files: [Upload!]!) { multipleUpload(files: $files)}",
			Variables: []byte(`{"files":[null, null]}`),
			Files:     files,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusRequestEntityTooLarge, res.Response.StatusCode)
		require.Equal(t, `{"errors":[{"message":"request body too large"}],"data":null}`, res.Body)
	})
}
