/*
 * Copyright 2021 - now, the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package zk

import (
	"github.com/go-zookeeper/zk"
	"github.com/monimesl/operator-helper/config"
	"strings"
)

const (
	// ClusterMetadataParentZNode defines the znode to store metadata for the ZookeeperCluster objects
	ClusterMetadataParentZNode = "/pulsar-operator-cluster-metadata"
)

type Client struct {
	conn                 *zk.Conn
	requiredNodesCreated bool
}

// Close closes the zookeeper connection
func (c *Client) Close() {
	config.RequireRootLogger().Info("Closing the zookeeper client")
	c.conn.Close()
}

func (c *Client) createRequiredNodes() (err error) {
	if !c.requiredNodesCreated {
		_ = c.setNodeData(ClusterMetadataParentZNode, nil)
	}
	return
}

func (c *Client) setNodeData(path string, data []byte) (err error) {
	config.RequireRootLogger().
		Info("Creating the operator metadata node",
			"path", path, "data", string(data))
	_, stats, err := c.getNode(path)
	if err == zk.ErrNoNode {
		return c.createNode(path, data)
	} else if err != nil {
		return err
	}
	if _, err = c.conn.Set(path, data, stats.Version); err == zk.ErrNodeExists {
		return nil
	}
	return
}

func (c *Client) getNode(clusterNode string) ([]byte, *zk.Stat, error) {
	data, sts, err := c.conn.Get(clusterNode)
	if err != nil {
		return nil, nil, err
	}
	return data, sts, nil
}

func (c *Client) createNode(path string, data []byte) error {
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	paths := strings.Split(path, "/")
	zNodes := make([]string, len(paths))
	for i := range paths {
		zNodes[i] = "/" + strings.Join(paths[0:i+1], "/")
	}
	for i, zNode := range zNodes {
		var nodeData []byte = nil
		if i == len(zNodes)-1 {
			// Last node
			nodeData = data
		}
		_, err := c.conn.Create(zNode, nodeData, 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return err
		}
	}
	return nil
}

func (c *Client) deleteNodes(paths ...string) error {
	for _, path := range paths {
		if err := c.deleteNode(path); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) deleteNode(path string) error {
	config.RequireRootLogger().
		Info("Deleting the zookeeper node",
			"zNode", path)
	_, stat, err := c.getNode(path)
	if err == zk.ErrNoNode {
		return nil
	} else if err != nil {
		return err
	}
	err = c.conn.Delete(path, stat.Version)
	if err == zk.ErrNotEmpty {
		children, err2 := c.getChildren(path)
		if err2 != nil {
			return err2
		}
		for i, child := range children {
			children[i] = path + "/" + child
		}
		err2 = c.deleteNodes(children...)
		if err2 != nil {
			return err2
		}
		return c.deleteNode(path)
	}
	return err
}

func (c *Client) getChildren(path string) ([]string, error) {
	children, _, err := c.conn.Children(path)
	if err != nil {
		return nil, err
	}
	return children, nil
}