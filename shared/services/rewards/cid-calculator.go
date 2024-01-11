package rewards

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	blockservice "github.com/ipfs/boxo/blockservice"
	blockstore "github.com/ipfs/boxo/blockstore"
	chunker "github.com/ipfs/boxo/chunker"
	merkledag "github.com/ipfs/boxo/ipld/merkledag"
	unixfs "github.com/ipfs/boxo/ipld/unixfs"
	balanced "github.com/ipfs/boxo/ipld/unixfs/importer/balanced"
	helpers "github.com/ipfs/boxo/ipld/unixfs/importer/helpers"
	mfs "github.com/ipfs/boxo/mfs"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
)

// Computes the CID for an arbitrary bytestring with a given filename
// by adding the file to an empty directory at the root level of the IPFS node.
//
// Only the last segment of the filename will be used, ie, `/home/alice/foo.zip`
// will be stripped to `foo.zip`.
func singleFileDirIPFSCid(data []byte, filename string) (cid.Cid, error) {
	ds := sync.MutexWrap(datastore.NewMapDatastore())
	bsvc := blockservice.New(blockstore.NewBlockstore(ds), nil)
	dag := merkledag.NewDAGService(bsvc)
	cidBuilder := merkledag.V1CidPrefix()

	// Strip the leading path segments to get the file name
	filename = filepath.Base(filename)

	// Create the root node, an empty directory
	rootNode := unixfs.EmptyDirNode()
	err := rootNode.SetCidBuilder(cidBuilder)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error creating the CID builder: %w", err)
	}
	root, err := mfs.NewRoot(context.Background(), dag, rootNode, nil)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error setting new MFS root: %w", err)
	}

	// Create a chunker-reader from the compressed data
	chnk, err := chunker.FromString(bytes.NewReader(data), "size-1048576")
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error creating chunker-reader from compressed bytes: %w", err)
	}
	// Create a DAG builder using the same settings as web3storage
	params := helpers.DagBuilderParams{
		Dagserv:    dag,
		RawLeaves:  true,
		Maxlinks:   1024,
		CidBuilder: cidBuilder,
	}
	ufsBuilder, err := params.New(chnk)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error creating params from chunk: %w", err)
	}

	// Create the node for the file in the DAG
	node, err := balanced.Layout(ufsBuilder)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error creating DAG layout: %w", err)
	}

	// Add the file to the root directory
	err = mfs.PutNode(root, filename, node)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error adding node to DAG: %w", err)
	}

	// Add the file to the dag
	_, err = mfs.NewFile(filename, node, nil, dag)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error adding compressed file to DAG: %w", err)
	}

	// Finalize the dag and get the cid
	rootDir := root.GetDirectory()

	err = rootDir.Flush()
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error flushing DAG root dir: %w", err)
	}

	err = root.Close()
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error closing DAG root: %w", err)
	}

	rootDirNode, err := rootDir.GetNode()
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error getting DAG root node: %w", err)
	}

	err = ufsBuilder.Add(rootDirNode)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error adding DAG root to UFS builder: %w", err)
	}
	return rootDirNode.Cid(), nil
}
