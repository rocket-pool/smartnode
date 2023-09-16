package server

import "github.com/rocket-pool/smartnode/shared/types/api"

type IContextFactory[Context ISingleStageCallContext[DataType, CommonContextType], DataType any, CommonContextType any] interface {
	Create(vars map[string]string) (Context, error)
	Run(c Context) (*api.ApiResponse[DataType], error)
}
