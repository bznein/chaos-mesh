// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package dbstore

import (
	"context"

	"go.uber.org/fx"

	"github.com/jinzhu/gorm"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	log = ctrl.Log.WithName("store/dbstore")
)

// DB defines a db storage.
type DB struct {
	*gorm.DB
}

// NewDBStore returns a new DB
func NewDBStore(lc fx.Lifecycle, conf *config.ChaosDashboardConfig) (*DB, error) {
	gormDB, err := gorm.Open(conf.Database.Driver, conf.Database.Datasource)
	if err != nil {
		log.Error(err, "failed to open DB")
		return nil, err
	}

	db := &DB{
		gormDB,
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return db.Close()
		},
	})

	return db, nil
}
