//go:generate mockery --name=UserInfoRepository --dir=../ --output=. --filename=user_info_repo_mock.go
//go:generate mockery --name=UserParamsRepository --dir=../ --output=. --filename=user_params_repo_mock.go
//go:generate mockery --name=UserWeightRepository --dir=../ --output=. --filename=user_weight_repo_mock.go
//go:generate mockery --name=TransactionManager --dir=../ --output=. --filename=trx_manager_mock.go
package mocks