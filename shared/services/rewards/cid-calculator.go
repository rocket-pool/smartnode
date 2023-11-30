package rewards

import (
	"bytes"
	"context"
	"fmt"
	"testing/fstest"
	"time"

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
	"github.com/web3-storage/go-w3s-client/adder"
)

func snCid(compressedData []byte, filename string) (cid.Cid, error) {

	// Create an in-memory file and FS
	mapFile := fstest.MapFile{
		Data:    compressedData,
		Mode:    0644,
		ModTime: time.Now(),
	}
	fsMap := fstest.MapFS{filename: &mapFile}
	file, err := fsMap.Open(filename)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error opening memory-mapped file: %w", err)
	}

	// Use the web3.storage libraries to chunk the data and get the root CID
	ds := sync.MutexWrap(datastore.NewMapDatastore())
	bsvc := blockservice.New(blockstore.NewBlockstore(ds), nil)
	dag := merkledag.NewDAGService(bsvc)
	dagFmtr, err := adder.NewAdder(context.Background(), dag)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error creating DAG adder: %w", err)
	}
	root, err := dagFmtr.Add(file, "", fsMap)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("error adding rewards file to DAG: %w", err)
	}

	return root, err
}

func SingleFileDirIPFSCid(data []byte, filename string) (cid.Cid, error) {

	ds := sync.MutexWrap(datastore.NewMapDatastore())
	bsvc := blockservice.New(blockstore.NewBlockstore(ds), nil)
	dag := merkledag.NewDAGService(bsvc)
	cidBuilder := merkledag.V1CidPrefix()

	// Create the root node, an empty directory
	rootNode := unixfs.EmptyDirNode()
	err := rootNode.SetCidBuilder(cidBuilder)
	if err != nil {
		return cid.Cid{}, err
	}
	root, err := mfs.NewRoot(context.Background(), dag, rootNode, nil)
	if err != nil {
		return cid.Cid{}, err
	}

	// Create a chunker-reader from the data
	chnk, err := chunker.FromString(bytes.NewReader(data), "size-1048576")
	if err != nil {
		return cid.Cid{}, err
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
		return cid.Cid{}, err
	}

	// Create the node for the file in the DAG
	node, err := balanced.Layout(ufsBuilder)
	if err != nil {
		return cid.Cid{}, err
	}

	// Add the file to the root directory
	err = mfs.PutNode(root, filename, node)
	if err != nil {
		return cid.Cid{}, err
	}

	// Add the file to the dag
	_, err = mfs.NewFile(filename, node, nil, dag)
	if err != nil {
		return cid.Cid{}, err
	}

	// Finalize the dag and get the cid
	rootDir := root.GetDirectory()

	err = rootDir.Flush()
	if err != nil {
		return cid.Cid{}, err
	}

	err = root.Close()
	if err != nil {
		return cid.Cid{}, err
	}

	rootDirNode, err := rootDir.GetNode()
	if err != nil {
		return cid.Cid{}, err
	}

	ufsBuilder.Add(rootDirNode)
	return rootDirNode.Cid(), nil
}
